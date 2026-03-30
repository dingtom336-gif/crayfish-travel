package lock

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xiaozhang/crayfish-travel/backend/internal/common/middleware"
)

// Handler handles lock-related HTTP requests.
type Handler struct {
	orchestrator *Orchestrator
}

// NewHandler creates a new lock handler.
func NewHandler(orchestrator *Orchestrator) *Handler {
	return &Handler{orchestrator: orchestrator}
}

// AcquireRequest is the request body to acquire a lock.
type AcquireRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	QuoteID   string `json:"quote_id" binding:"required"`
}

// AcquireResponse contains the lock acquisition result.
type AcquireResponse struct {
	LockSessionID string `json:"lock_session_id"`
	State         string `json:"state"`
	ExpiresAt     string `json:"expires_at,omitempty"`
	TraceID       string `json:"trace_id"`
}

// Acquire godoc
// @Summary      Acquire a lock on a quote
// @Description  Run saga to lock a supplier quote for 15 minutes, freezing funds
// @Tags         lock
// @Accept       json
// @Produce      json
// @Param        body  body  AcquireRequest  true  "Session and Quote IDs"
// @Success      200   {object}  AcquireResponse
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /lock/acquire [post]
func (h *Handler) Acquire(c *gin.Context) {
	var req AcquireRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	traceID := middleware.GetTraceID(c)

	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session_id"})
		return
	}

	quoteID, err := uuid.Parse(req.QuoteID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid quote_id"})
		return
	}

	ls, err := h.orchestrator.AcquireLock(c.Request.Context(), sessionID, quoteID, traceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "trace_id": traceID})
		return
	}

	resp := AcquireResponse{
		LockSessionID: ls.ID.String(),
		State:         ls.State,
		TraceID:       traceID,
	}
	if ls.ExpiresAt != nil {
		resp.ExpiresAt = ls.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
	}

	c.JSON(http.StatusOK, resp)
}

// StatusResponse contains the lock status and remaining TTL.
type StatusResponse struct {
	LockSessionID string `json:"lock_session_id"`
	SessionID     string `json:"session_id"`
	QuoteID       string `json:"quote_id"`
	State         string `json:"state"`
	RemainingTTL  int    `json:"remaining_ttl_seconds"`
	TraceID       string `json:"trace_id"`
}

// Status godoc
// @Summary      Get lock status for a session
// @Description  Returns lock state and remaining TTL in seconds
// @Tags         lock
// @Produce      json
// @Param        session_id  path  string  true  "Session ID"
// @Success      200  {object}  StatusResponse
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /lock/{session_id}/status [get]
func (h *Handler) Status(c *gin.Context) {
	sessionIDStr := c.Param("session_id")
	traceID := middleware.GetTraceID(c)

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session_id"})
		return
	}

	ls, err := h.orchestrator.GetLockBySessionID(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no lock found for session", "trace_id": traceID})
		return
	}

	var remainingSeconds int
	if ls.State == StateLocked {
		ttl, err := h.orchestrator.GetRemainingTTL(c.Request.Context(), ls.SessionID, ls.QuoteID)
		if err == nil && ttl > 0 {
			remainingSeconds = int(ttl.Seconds())
		}
	}

	c.JSON(http.StatusOK, StatusResponse{
		LockSessionID: ls.ID.String(),
		SessionID:     ls.SessionID.String(),
		QuoteID:       ls.QuoteID.String(),
		State:         ls.State,
		RemainingTTL:  remainingSeconds,
		TraceID:       traceID,
	})
}
