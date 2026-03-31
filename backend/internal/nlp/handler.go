package nlp

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xiaozhang/crayfish-travel/backend/internal/common/middleware"
)

// NLPParser defines the interface for parsing natural language into travel requirements.
type NLPParser interface {
	Parse(rawInput string) (*TravelRequirement, error)
}

// FallbackParser tries ARKClient first, then HeuristicParser as fallback.
type FallbackParser struct {
	ark       *ARKClient
	heuristic *HeuristicParser
}

// NewFallbackParser creates a parser that tries ARK first, heuristic second.
func NewFallbackParser(ark *ARKClient, heuristic *HeuristicParser) *FallbackParser {
	return &FallbackParser{ark: ark, heuristic: heuristic}
}

// Parse tries ARK first; on failure falls back to heuristic extraction.
func (f *FallbackParser) Parse(rawInput string) (*TravelRequirement, error) {
	// Try ARK LLM first
	result, err := f.ark.Parse(rawInput)
	if err == nil {
		return result, nil
	}

	// Fall back to heuristic parser
	hResult, matched := f.heuristic.Parse(rawInput)
	if matched {
		return hResult, nil
	}

	return nil, fmt.Errorf("all parsers failed: ark error: %w; heuristic: no match", err)
}

// Handler handles NLP-related HTTP requests.
type Handler struct {
	parser    NLPParser
	validator *DateValidator
	db        *sql.DB
}

// NewHandler creates a new NLP handler.
func NewHandler(db *sql.DB, parser NLPParser, validator *DateValidator) *Handler {
	return &Handler{
		parser:    parser,
		validator: validator,
		db:        db,
	}
}

// ParseRequest is the request body for NLP parsing.
type ParseRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	RawInput  string `json:"raw_input" binding:"required"`
}

// ParseResponse contains the parsed travel requirements.
type ParseResponse struct {
	SessionID   string             `json:"session_id"`
	Requirement *TravelRequirement `json:"requirement"`
	Validation  *DateValidation    `json:"validation"`
	TraceID     string             `json:"trace_id"`
}

// Parse godoc
// @Summary      Parse travel requirements
// @Description  Use Claude AI to parse natural language travel input into structured data
// @Tags         nlp
// @Accept       json
// @Produce      json
// @Param        body  body  ParseRequest  true  "Natural language input"
// @Success      200   {object}  ParseResponse
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /nlp/parse [post]
func (h *Handler) Parse(c *gin.Context) {
	var req ParseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	traceID := middleware.GetTraceID(c)

	// Parse with NLP parser (ARK + heuristic fallback)
	requirement, err := h.parser.Parse(req.RawInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse requirements"})
		return
	}

	// Validate dates
	validation, err := h.validator.Validate(requirement.StartDate, requirement.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update session with raw input
	_, err = h.db.Exec(`
		UPDATE sessions SET raw_input = $1, updated_at = NOW() WHERE id = $2`,
		req.RawInput, req.SessionID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update session"})
		return
	}

	c.JSON(http.StatusOK, ParseResponse{
		SessionID:   req.SessionID,
		Requirement: requirement,
		Validation:  validation,
		TraceID:     traceID,
	})
}

// ConfirmRequest is the request body for confirming parsed requirements.
type ConfirmRequest struct {
	SessionID   string   `json:"session_id" binding:"required"`
	Destination string   `json:"destination" binding:"required"`
	StartDate   string   `json:"start_date" binding:"required"`
	EndDate     string   `json:"end_date" binding:"required"`
	BudgetCents int64    `json:"budget_cents" binding:"required"`
	Adults      int      `json:"adults" binding:"required,min=1"`
	Children    int      `json:"children" binding:"min=0"`
	Preferences []string `json:"preferences"`
}

// ConfirmResponse is the response after confirming requirements.
type ConfirmResponse struct {
	SessionID    string `json:"session_id"`
	Status       string `json:"status"`
	IsPeakSeason bool   `json:"is_peak_season"`
	PeakType     string `json:"peak_type,omitempty"`
	TraceID      string `json:"trace_id"`
}

// Confirm godoc
// @Summary      Confirm parsed requirements
// @Description  User confirms the AI-parsed travel requirements, updating the session
// @Tags         nlp
// @Accept       json
// @Produce      json
// @Param        body  body  ConfirmRequest  true  "Confirmed requirements"
// @Success      200   {object}  ConfirmResponse
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /nlp/confirm [post]
func (h *Handler) Confirm(c *gin.Context) {
	var req ConfirmRequest
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

	// Validate dates for peak season
	validation, err := h.validator.Validate(req.StartDate, req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update session with confirmed requirements
	prefsJSON, _ := json.Marshal(req.Preferences)
	_, err = h.db.Exec(`
		UPDATE sessions SET
			destination = $1, start_date = $2, end_date = $3,
			budget_cents = $4, adults = $5, children = $6,
			preferences = $7, is_peak_season = $8, peak_type = $9,
			status = 'requirements_confirmed', updated_at = NOW()
		WHERE id = $10`,
		req.Destination, req.StartDate, req.EndDate,
		req.BudgetCents, req.Adults, req.Children,
		prefsJSON, validation.IsPeakSeason, validation.PeakType,
		sessionID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update session"})
		return
	}

	c.JSON(http.StatusOK, ConfirmResponse{
		SessionID:    req.SessionID,
		Status:       "requirements_confirmed",
		IsPeakSeason: validation.IsPeakSeason,
		PeakType:     validation.PeakType,
		TraceID:      traceID,
	})
}
