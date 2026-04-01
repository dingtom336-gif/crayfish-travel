package aiparser

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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

// Parse tries heuristic first (instant); only calls ARK LLM when heuristic can't match.
func (f *FallbackParser) Parse(rawInput string) (*TravelRequirement, error) {
	// Try heuristic first (0ms, no network)
	hResult, matched := f.heuristic.Parse(rawInput)
	if matched {
		return hResult, nil
	}

	// Heuristic couldn't match -- fall back to ARK LLM (slower but smarter)
	result, err := f.ark.Parse(rawInput)
	if err == nil && result.Destination != "" {
		return result, nil
	}

	return nil, fmt.Errorf("无法解析旅行需求，请描述更具体的目的地")
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无法识别您的旅行目的地，请描述更具体，例如：想去三亚玩5天",
		})
		return
	}

	if requirement.Destination == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请告诉我们您想去哪里，例如：去三亚、去泰国",
		})
		return
	}

	// Validate dates only if the parser extracted them
	var validation *DateValidation
	if requirement.StartDate != "" && requirement.EndDate != "" {
		validation, err = h.validator.Validate(requirement.StartDate, requirement.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		validation = &DateValidation{IsValid: true}
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

// ParseSSE streams parsing progress via Server-Sent Events.
func (h *Handler) ParseSSE(c *gin.Context) {
	var req ParseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Flush()

	sendEvent := func(event, data string) {
		fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, data)
		c.Writer.Flush()
	}

	// Step 1: parsing started
	sendEvent("progress", `{"step":"parsing","message":"正在解析您的需求..."}`)

	// Step 2: actual parse
	requirement, err := h.parser.Parse(req.RawInput)
	if err != nil {
		sendEvent("error", `{"message":"无法识别您的旅行目的地，请描述更具体，例如：想去三亚玩5天"}`)
		return
	}

	if requirement.Destination == "" {
		sendEvent("error", `{"message":"请告诉我们您想去哪里，例如：去三亚、去泰国"}`)
		return
	}

	sendEvent("progress", fmt.Sprintf(`{"step":"destination","message":"已识别目的地：%s"}`, requirement.Destination))

	// Step 3: budget
	if requirement.BudgetCents > 0 {
		sendEvent("progress", fmt.Sprintf(`{"step":"budget","message":"预算：%d元"}`, requirement.BudgetCents/100))
	} else {
		sendEvent("progress", `{"step":"budget","message":"将为您智能推荐预算范围"}`)
	}

	// Step 4: preferences
	if len(requirement.Preferences) > 0 {
		sendEvent("progress", fmt.Sprintf(`{"step":"preferences","message":"偏好：%s"}`, strings.Join(requirement.Preferences, "、")))
	}

	// Step 5: date validation (if dates provided)
	var validation *DateValidation
	if requirement.StartDate != "" && requirement.EndDate != "" {
		validation, err = h.validator.Validate(requirement.StartDate, requirement.EndDate)
		if err != nil {
			sendEvent("error", fmt.Sprintf(`{"message":"%s"}`, err.Error()))
			return
		}
		if validation.IsPeakSeason {
			sendEvent("progress", `{"step":"peak","message":"注意：高峰期价格可能上浮"}`)
		}
	} else {
		validation = &DateValidation{IsValid: true}
	}

	// Step 6: update session
	if req.SessionID != "" {
		h.db.Exec(`UPDATE sessions SET raw_input = $1, updated_at = NOW() WHERE id = $2`, req.RawInput, req.SessionID)
	}

	// Step 7: send final result
	resultJSON, _ := json.Marshal(ParseResponse{
		SessionID:   req.SessionID,
		Requirement: requirement,
		Validation:  validation,
		TraceID:     middleware.GetTraceID(c),
	})
	sendEvent("result", string(resultJSON))
	sendEvent("done", `{}`)
}

// ConfirmRequest is the request body for confirming parsed requirements.
type ConfirmRequest struct {
	SessionID   string   `json:"session_id" binding:"required"`
	Destination string   `json:"destination" binding:"required"`
	StartDate   string   `json:"start_date" binding:"required"`
	EndDate     string   `json:"end_date" binding:"required"`
	BudgetCents int64    `json:"budget_cents" binding:"min=0"`
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

	// Validate required date fields
	if req.StartDate == "" || req.EndDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请填写出发日期和返回日期"})
		return
	}

	// Validate dates for peak season
	validation, err := h.validator.Validate(req.StartDate, req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update session with confirmed requirements
	// Ensure preferences is never null in JSON output
	prefs := req.Preferences
	if prefs == nil {
		prefs = []string{}
	}
	prefsJSON, _ := json.Marshal(prefs)
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
