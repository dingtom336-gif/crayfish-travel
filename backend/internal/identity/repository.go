package identity

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Record represents an identity record with encrypted PII.
type Record struct {
	ID              uuid.UUID
	SessionID       uuid.UUID
	TraceID         string
	NameEncrypted   []byte
	IDNumEncrypted  []byte
	PhoneEncrypted  []byte
	Nonce           []byte
	ExpiresAt       time.Time
	CreatedAt       time.Time
}

// Repository handles identity_records CRUD operations.
type Repository struct {
	db        *sql.DB
	encryptor *Encryptor
}

// NewRepository creates a new identity repository.
func NewRepository(db *sql.DB, encryptor *Encryptor) *Repository {
	return &Repository{db: db, encryptor: encryptor}
}

// CreateInput holds the raw PII for creating an identity record.
type CreateInput struct {
	SessionID uuid.UUID
	TraceID   string
	Name      string
	IDNumber  string
	Phone     string
	TTL       time.Duration
}

// Create encrypts PII and stores in the database.
// Each field is sealed with its own nonce (prepended to ciphertext).
// The legacy `nonce` column stores an empty placeholder for schema compatibility.
func (r *Repository) Create(input CreateInput) (*Record, error) {
	nameEnc, err := r.encryptor.SealStringWithNonce(input.Name)
	if err != nil {
		return nil, fmt.Errorf("encrypt name: %w", err)
	}

	idEnc, err := r.encryptor.SealStringWithNonce(input.IDNumber)
	if err != nil {
		return nil, fmt.Errorf("encrypt id: %w", err)
	}

	phoneEnc, err := r.encryptor.SealStringWithNonce(input.Phone)
	if err != nil {
		return nil, fmt.Errorf("encrypt phone: %w", err)
	}

	expiresAt := time.Now().Add(input.TTL)
	emptyNonce := []byte{} // legacy column placeholder

	var id uuid.UUID
	err = r.db.QueryRow(`
		INSERT INTO identity_records (session_id, trace_id, name_encrypted, id_number_encrypted, phone_encrypted, nonce, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`,
		input.SessionID, input.TraceID, nameEnc, idEnc, phoneEnc, emptyNonce, expiresAt,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("insert identity: %w", err)
	}

	return &Record{
		ID:        id,
		SessionID: input.SessionID,
		TraceID:   input.TraceID,
		ExpiresAt: expiresAt,
	}, nil
}

// FindBySessionID retrieves identity record by session ID.
func (r *Repository) FindBySessionID(sessionID uuid.UUID) (*Record, error) {
	rec := &Record{}
	err := r.db.QueryRow(`
		SELECT id, session_id, trace_id, name_encrypted, id_number_encrypted, phone_encrypted, nonce, expires_at, created_at
		FROM identity_records
		WHERE session_id = $1`,
		sessionID,
	).Scan(&rec.ID, &rec.SessionID, &rec.TraceID, &rec.NameEncrypted, &rec.IDNumEncrypted, &rec.PhoneEncrypted, &rec.Nonce, &rec.ExpiresAt, &rec.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("find identity: %w", err)
	}
	return rec, nil
}

// DeleteBySessionID removes identity record for a session.
func (r *Repository) DeleteBySessionID(sessionID uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM identity_records WHERE session_id = $1`, sessionID)
	if err != nil {
		return fmt.Errorf("delete identity: %w", err)
	}
	return nil
}

// DeleteExpired removes all records past their TTL.
func (r *Repository) DeleteExpired() (int64, error) {
	result, err := r.db.Exec(`DELETE FROM identity_records WHERE expires_at < NOW()`)
	if err != nil {
		return 0, fmt.Errorf("delete expired: %w", err)
	}
	return result.RowsAffected()
}
