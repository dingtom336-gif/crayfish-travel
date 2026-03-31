package identity

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xiaozhang/crayfish-travel/backend/internal/common/middleware"
)

// Handler handles identity-related HTTP requests.
type Handler struct {
	repo *Repository
	db   *sql.DB
}

// NewHandler creates a new identity handler.
func NewHandler(db *sql.DB, encryptor *Encryptor) *Handler {
	return &Handler{
		repo: NewRepository(db, encryptor),
		db:   db,
	}
}

// CreateRequest is the request body for identity creation.
type CreateRequest struct {
	Name     string `json:"name" binding:"required"`
	IDNumber string `json:"id_number" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	Adults   int    `json:"adults" binding:"required,min=1"`
	Children int    `json:"children" binding:"min=0"`
}

// CreateResponse is the response after identity creation.
type CreateResponse struct {
	SessionID string `json:"session_id"`
	ExpiresAt string `json:"expires_at"`
	TraceID   string `json:"trace_id"`
}

// Create godoc
// @Summary      Create identity record
// @Description  Collect and encrypt traveler PII, create session
// @Tags         identity
// @Accept       json
// @Produce      json
// @Param        body  body  CreateRequest  true  "Traveler identity"
// @Success      201   {object}  CreateResponse
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /identity [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	traceID := middleware.GetTraceID(c)

	// Create session first
	var sessionID uuid.UUID
	err := h.db.QueryRow(`
		INSERT INTO sessions (trace_id, status, adults, children)
		VALUES ($1, 'identity_collected', $2, $3)
		RETURNING id`,
		traceID, req.Adults, req.Children,
	).Scan(&sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	// Encrypt and store PII
	rec, err := h.repo.Create(CreateInput{
		SessionID: sessionID,
		TraceID:   traceID,
		Name:      req.Name,
		IDNumber:  req.IDNumber,
		Phone:     req.Phone,
		TTL:       72 * time.Hour,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store identity"})
		return
	}

	c.JSON(http.StatusCreated, CreateResponse{
		SessionID: sessionID.String(),
		ExpiresAt: rec.ExpiresAt.Format(time.RFC3339),
		TraceID:   traceID,
	})
}

// CreateSessionRequest is the request for lightweight session creation (no PII).
type CreateSessionRequest struct {
	Adults   int `json:"adults" binding:"required,min=1"`
	Children int `json:"children" binding:"min=0"`
}

// CreateSession creates a session without PII collection.
func (h *Handler) CreateSession(c *gin.Context) {
	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	traceID := middleware.GetTraceID(c)
	var sessionID uuid.UUID
	err := h.db.QueryRow(`
		INSERT INTO sessions (trace_id, status, adults, children)
		VALUES ($1, 'session_created', $2, $3)
		RETURNING id`,
		traceID, req.Adults, req.Children,
	).Scan(&sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"session_id": sessionID.String(),
		"trace_id":   traceID,
	})
}

// GetSession returns the current session state.
func (h *Handler) GetSession(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session_id"})
		return
	}

	var dest, startDate, endDate, status sql.NullString
	var budgetCents sql.NullInt64
	var adults, children sql.NullInt32
	var prefsJSON sql.NullString

	err = h.db.QueryRow(`
		SELECT destination, start_date, end_date, budget_cents, adults, children, preferences, status
		FROM sessions WHERE id = $1`, sessionID,
	).Scan(&dest, &startDate, &endDate, &budgetCents, &adults, &children, &prefsJSON, &status)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	var prefs []string
	if prefsJSON.Valid && prefsJSON.String != "" {
		_ = json.Unmarshal([]byte(prefsJSON.String), &prefs)
	}
	if prefs == nil {
		prefs = []string{}
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id":  sessionID.String(),
		"destination": dest.String,
		"start_date":  startDate.String,
		"end_date":    endDate.String,
		"budget_cents": budgetCents.Int64,
		"adults":      adults.Int32,
		"children":    children.Int32,
		"preferences": prefs,
		"status":      status.String,
	})
}

// Delete godoc
// @Summary      Delete identity record
// @Description  Remove PII for a session (manual cleanup or TTL trigger)
// @Tags         identity
// @Param        session_id  path  string  true  "Session ID"
// @Success      204
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /identity/{session_id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	sessionIDStr := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session_id"})
		return
	}

	if err := h.repo.DeleteBySessionID(sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete identity"})
		return
	}

	c.Status(http.StatusNoContent)
}
