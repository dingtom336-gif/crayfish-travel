package lock

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"github.com/xiaozhang/crayfish-travel/backend/internal/riskcontrol"
)

// Saga states
const (
	StatePending      = "PENDING"
	StateLocking      = "LOCKING"
	StateLocked       = "LOCKED"
	StatePaying       = "PAYING"
	StatePaid         = "PAID"
	StateCompleted    = "COMPLETED"
	StateLockFailed   = "LOCK_FAILED"
	StateReleasing    = "RELEASING"
	StateReleased     = "RELEASED"
	StateCompensating = "COMPENSATING"
	StateCompensated  = "COMPENSATED"
)

// Lock TTL: 15 minutes
const lockTTL = 15 * time.Minute

// LockSession represents a lock session record.
type LockSession struct {
	ID         uuid.UUID  `json:"id"`
	SessionID  uuid.UUID  `json:"session_id"`
	TraceID    string     `json:"trace_id"`
	QuoteID    uuid.UUID  `json:"quote_id"`
	State      string     `json:"state"`
	LockedAt   *time.Time `json:"locked_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	ReleasedAt *time.Time `json:"released_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// Orchestrator manages the Saga state machine for lock acquisition.
type Orchestrator struct {
	db   *sql.DB
	rdb  *goredis.Client
	pool *riskcontrol.FundPool
}

// NewOrchestrator creates a new lock Orchestrator.
func NewOrchestrator(db *sql.DB, rdb *goredis.Client, pool *riskcontrol.FundPool) *Orchestrator {
	return &Orchestrator{db: db, rdb: rdb, pool: pool}
}

// AcquireLock runs the forward saga to lock a quote for a session.
func (o *Orchestrator) AcquireLock(ctx context.Context, sessionID, quoteID uuid.UUID, traceID string) (*LockSession, error) {
	// Step 0: Create lock_session in PENDING state
	lockID := uuid.New()
	_, err := o.db.Exec(`
		INSERT INTO lock_sessions (id, session_id, trace_id, quote_id, state)
		VALUES ($1, $2, $3, $4, $5)`,
		lockID, sessionID, traceID, quoteID, StatePending,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create lock session: %w", err)
	}

	// Transition to LOCKING
	if err := o.updateState(lockID, StateLocking); err != nil {
		return nil, err
	}

	// Forward saga steps
	steps := []struct {
		name    string
		order   int
		execute func() error
	}{
		{
			name:  "lock_supplier",
			order: 1,
			execute: func() error {
				// Mock supplier lock - in production this calls the supplier API
				log.Printf("[%s] saga: locking supplier for quote %s", traceID, quoteID)
				return nil
			},
		},
		{
			name:  "freeze_funds",
			order: 2,
			execute: func() error {
				// Look up quote price to freeze
				var totalCents int64
				err := o.db.QueryRow(
					`SELECT total_price_cents FROM supplier_quotes WHERE id = $1`, quoteID,
				).Scan(&totalCents)
				if err != nil {
					return fmt.Errorf("failed to get quote price: %w", err)
				}
				return o.pool.Freeze(sessionID, totalCents, traceID)
			},
		},
		{
			name:  "confirm_lock",
			order: 3,
			execute: func() error {
				// Set Redis lock key with 15min TTL
				redisKey := lockRedisKey(sessionID, quoteID)
				return o.rdb.Set(ctx, redisKey, lockID.String(), lockTTL).Err()
			},
		},
	}

	// Execute forward steps, compensate on failure
	completedSteps := 0
	for _, step := range steps {
		if err := o.logSagaStep(lockID, traceID, step.name, step.order, "forward", "running"); err != nil {
			log.Printf("[%s] saga: failed to log step %s: %v", traceID, step.name, err)
		}

		if err := step.execute(); err != nil {
			// Log failure
			o.completeSagaStep(lockID, step.name, "forward", "failed", err.Error())
			// Run compensation
			o.compensate(ctx, lockID, sessionID, quoteID, traceID, steps[:completedSteps])
			return nil, fmt.Errorf("saga step %s failed: %w", step.name, err)
		}

		o.completeSagaStep(lockID, step.name, "forward", "completed", "")
		completedSteps++
	}

	// All steps succeeded - transition to LOCKED
	now := time.Now()
	expiresAt := now.Add(lockTTL)
	_, err = o.db.Exec(`
		UPDATE lock_sessions SET state = $1, locked_at = $2, expires_at = $3, updated_at = NOW()
		WHERE id = $4`,
		StateLocked, now, expiresAt, lockID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update lock state: %w", err)
	}

	return &LockSession{
		ID:        lockID,
		SessionID: sessionID,
		TraceID:   traceID,
		QuoteID:   quoteID,
		State:     StateLocked,
		LockedAt:  &now,
		ExpiresAt: &expiresAt,
	}, nil
}

// ReleaseLock releases an existing lock and unfreezes funds.
func (o *Orchestrator) ReleaseLock(ctx context.Context, lockSessionID uuid.UUID, traceID string) error {
	// Load the lock session
	var sessionID, quoteID uuid.UUID
	var state string
	err := o.db.QueryRow(`
		SELECT session_id, quote_id, state FROM lock_sessions WHERE id = $1`,
		lockSessionID,
	).Scan(&sessionID, &quoteID, &state)
	if err != nil {
		return fmt.Errorf("failed to find lock session: %w", err)
	}

	if state == StateReleased || state == StateCompensated {
		return nil // already released
	}

	// Transition to RELEASING
	if err := o.updateState(lockSessionID, StateReleasing); err != nil {
		return err
	}

	// Delete Redis key
	redisKey := lockRedisKey(sessionID, quoteID)
	o.rdb.Del(ctx, redisKey)

	// Unfreeze funds
	if err := o.pool.Unfreeze(sessionID, traceID); err != nil {
		log.Printf("[%s] release: unfreeze failed (may already be unfrozen): %v", traceID, err)
	}

	// Transition to RELEASED
	now := time.Now()
	_, err = o.db.Exec(`
		UPDATE lock_sessions SET state = $1, released_at = $2, updated_at = NOW()
		WHERE id = $3`,
		StateReleased, now, lockSessionID,
	)
	if err != nil {
		return fmt.Errorf("failed to update lock to released: %w", err)
	}

	return nil
}

// GetLockSession retrieves a lock session by ID.
func (o *Orchestrator) GetLockSession(lockSessionID uuid.UUID) (*LockSession, error) {
	var ls LockSession
	err := o.db.QueryRow(`
		SELECT id, session_id, trace_id, quote_id, state,
		       locked_at, expires_at, released_at, created_at, updated_at
		FROM lock_sessions WHERE id = $1`,
		lockSessionID,
	).Scan(&ls.ID, &ls.SessionID, &ls.TraceID, &ls.QuoteID, &ls.State,
		&ls.LockedAt, &ls.ExpiresAt, &ls.ReleasedAt, &ls.CreatedAt, &ls.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get lock session: %w", err)
	}
	return &ls, nil
}

// GetLockBySessionID retrieves the latest lock session for a given session.
func (o *Orchestrator) GetLockBySessionID(sessionID uuid.UUID) (*LockSession, error) {
	var ls LockSession
	err := o.db.QueryRow(`
		SELECT id, session_id, trace_id, quote_id, state,
		       locked_at, expires_at, released_at, created_at, updated_at
		FROM lock_sessions WHERE session_id = $1
		ORDER BY created_at DESC LIMIT 1`,
		sessionID,
	).Scan(&ls.ID, &ls.SessionID, &ls.TraceID, &ls.QuoteID, &ls.State,
		&ls.LockedAt, &ls.ExpiresAt, &ls.ReleasedAt, &ls.CreatedAt, &ls.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get lock session by session_id: %w", err)
	}
	return &ls, nil
}

// GetRemainingTTL returns the remaining TTL for a lock's Redis key.
func (o *Orchestrator) GetRemainingTTL(ctx context.Context, sessionID, quoteID uuid.UUID) (time.Duration, error) {
	redisKey := lockRedisKey(sessionID, quoteID)
	ttl, err := o.rdb.TTL(ctx, redisKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}
	return ttl, nil
}

// compensate runs compensation steps in reverse order.
func (o *Orchestrator) compensate(ctx context.Context, lockID, sessionID, quoteID uuid.UUID, traceID string, completedSteps []struct {
	name    string
	order   int
	execute func() error
}) {
	o.updateState(lockID, StateCompensating)

	for i := len(completedSteps) - 1; i >= 0; i-- {
		step := completedSteps[i]
		compName := "compensate_" + step.name

		o.logSagaStep(lockID, traceID, compName, step.order, "compensate", "running")

		var compErr error
		switch step.name {
		case "confirm_lock":
			redisKey := lockRedisKey(sessionID, quoteID)
			compErr = o.rdb.Del(ctx, redisKey).Err()
		case "freeze_funds":
			compErr = o.pool.Unfreeze(sessionID, traceID)
		case "lock_supplier":
			// Mock: nothing to compensate
			log.Printf("[%s] saga: compensate supplier lock (mock)", traceID)
		}

		if compErr != nil {
			o.completeSagaStep(lockID, compName, "compensate", "failed", compErr.Error())
			log.Printf("[%s] saga: compensation step %s failed: %v", traceID, compName, compErr)
		} else {
			o.completeSagaStep(lockID, compName, "compensate", "completed", "")
		}
	}

	o.updateState(lockID, StateLockFailed)
}

// updateState transitions the lock session to a new state.
func (o *Orchestrator) updateState(lockID uuid.UUID, state string) error {
	_, err := o.db.Exec(
		`UPDATE lock_sessions SET state = $1, updated_at = NOW() WHERE id = $2`,
		state, lockID,
	)
	if err != nil {
		return fmt.Errorf("failed to update state to %s: %w", state, err)
	}
	return nil
}

// logSagaStep inserts a saga step record.
func (o *Orchestrator) logSagaStep(lockID uuid.UUID, traceID, stepName string, stepOrder int, direction, status string) error {
	_, err := o.db.Exec(`
		INSERT INTO saga_steps (lock_session_id, trace_id, step_name, step_order, direction, status, started_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())`,
		lockID, traceID, stepName, stepOrder, direction, status,
	)
	return err
}

// completeSagaStep updates a saga step's final status.
func (o *Orchestrator) completeSagaStep(lockID uuid.UUID, stepName, direction, status, errMsg string) {
	o.db.Exec(`
		UPDATE saga_steps SET status = $1, error_message = $2, completed_at = NOW()
		WHERE lock_session_id = $3 AND step_name = $4 AND direction = $5 AND completed_at IS NULL`,
		status, errMsg, lockID, stepName, direction,
	)
}

// lockRedisKey builds the Redis key for a lock.
func lockRedisKey(sessionID, quoteID uuid.UUID) string {
	return fmt.Sprintf("lock:%s:%s", sessionID, quoteID)
}
