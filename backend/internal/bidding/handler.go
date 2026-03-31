package bidding

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xiaozhang/crayfish-travel/backend/internal/common/middleware"
)

// Handler handles bidding-related HTTP requests.
type Handler struct {
	supplier SupplierClient
	db       *sql.DB
}

// NewHandler creates a new bidding handler.
func NewHandler(db *sql.DB, supplier SupplierClient) *Handler {
	return &Handler{supplier: supplier, db: db}
}

// StartRequest is the request body to start a bidding session.
type StartRequest struct {
	SessionID string `json:"session_id" binding:"required"`
}

// StartResponse contains the bidding results.
type StartResponse struct {
	SessionID string        `json:"session_id"`
	Packages  []RankedQuote `json:"packages"`
	Count     int           `json:"count"`
	TraceID   string        `json:"trace_id"`
}

// Start godoc
// @Summary      Start bidding for travel packages
// @Description  Fetch supplier quotes, rank top 5, and return with mandatory price breakdown
// @Tags         bidding
// @Accept       json
// @Produce      json
// @Param        body  body  StartRequest  true  "Session ID"
// @Success      200   {object}  StartResponse
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /bidding/start [post]
func (h *Handler) Start(c *gin.Context) {
	var req StartRequest
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

	// Load session to get requirements
	var dest string
	var budgetCents int64
	var adults, children int
	var startDate, endDate string
	err = h.db.QueryRow(`
		SELECT destination, budget_cents, adults, children, start_date, end_date
		FROM sessions WHERE id = $1 AND status = 'requirements_confirmed'`,
		sessionID,
	).Scan(&dest, &budgetCents, &adults, &children, &startDate, &endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session not found or requirements not confirmed"})
		return
	}

	// Calculate duration from actual dates
	startTime, _ := time.Parse("2006-01-02", startDate)
	endTime, _ := time.Parse("2006-01-02", endDate)
	days := int(endTime.Sub(startTime).Hours() / 24)
	if days < 1 {
		days = 1
	}

	// Fetch quotes from supplier
	quotes, err := h.supplier.FetchQuotes(dest, days, budgetCents, adults, children)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch packages"})
		return
	}

	// Rank top 5
	ranked := RankTop5(quotes)

	// Validate price split (compliance)
	if idx := ValidatePriceSplit(ranked); idx >= 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "price split validation failed"})
		return
	}

	// Persist quotes to database and capture IDs
	for i, q := range ranked {
		highlightsJSON, _ := json.Marshal(q.Highlights)
		inclusionsJSON, _ := json.Marshal(q.Inclusions)

		var quoteID string
		err = h.db.QueryRow(`
			INSERT INTO supplier_quotes
				(session_id, trace_id, supplier, package_title, destination,
				 duration_days, duration_nights, base_price_cents, refund_guarantee_fee_cents,
				 total_price_cents, star_rating, review_count, hotel_name,
				 highlights, inclusions, image_url, rank, is_best_value)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
			RETURNING id`,
			sessionID, traceID, q.Supplier, q.PackageTitle, q.Destination,
			q.DurationDays, q.DurationNights, q.BasePriceCents, q.RefundGuaranteeFeeCents,
			q.TotalPriceCents, q.StarRating, q.ReviewCount, q.HotelName,
			highlightsJSON, inclusionsJSON, q.ImageURL, q.Rank, q.IsBestValue,
		).Scan(&quoteID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save quotes"})
			return
		}
		ranked[i].ID = quoteID
	}

	// Update session status
	if _, err := h.db.Exec(`UPDATE sessions SET status = 'bidding_complete', updated_at = NOW() WHERE id = $1`, sessionID); err != nil {
		log.Printf("[bidding:start] failed to update session %s status: %v", sessionID, err)
	}

	c.JSON(http.StatusOK, StartResponse{
		SessionID: req.SessionID,
		Packages:  ranked,
		Count:     len(ranked),
		TraceID:   traceID,
	})
}

// PackagesResponse lists saved packages for a session.
type PackagesResponse struct {
	SessionID string        `json:"session_id"`
	Packages  []RankedQuote `json:"packages"`
	TraceID   string        `json:"trace_id"`
}

// GetPackages godoc
// @Summary      Get packages for a session
// @Description  Retrieve previously fetched and ranked packages
// @Tags         bidding
// @Produce      json
// @Param        session_id  path  string  true  "Session ID"
// @Success      200  {object}  PackagesResponse
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /bidding/{session_id}/packages [get]
func (h *Handler) GetPackages(c *gin.Context) {
	sessionIDStr := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session_id"})
		return
	}

	traceID := middleware.GetTraceID(c)

	rows, err := h.db.Query(`
		SELECT id, supplier, package_title, destination, duration_days, duration_nights,
		       base_price_cents, refund_guarantee_fee_cents, total_price_cents,
		       star_rating, review_count, hotel_name, highlights, inclusions,
		       image_url, rank, is_best_value
		FROM supplier_quotes
		WHERE session_id = $1
		ORDER BY rank ASC`,
		sessionID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load packages"})
		return
	}
	defer rows.Close()

	var packages []RankedQuote
	for rows.Next() {
		var q RankedQuote
		var highlightsJSON, inclusionsJSON []byte
		err := rows.Scan(
			&q.ID, &q.Supplier, &q.PackageTitle, &q.Destination,
			&q.DurationDays, &q.DurationNights,
			&q.BasePriceCents, &q.RefundGuaranteeFeeCents, &q.TotalPriceCents,
			&q.StarRating, &q.ReviewCount, &q.HotelName,
			&highlightsJSON, &inclusionsJSON,
			&q.ImageURL, &q.Rank, &q.IsBestValue,
		)
		if err != nil {
			continue
		}
		json.Unmarshal(highlightsJSON, &q.Highlights)
		json.Unmarshal(inclusionsJSON, &q.Inclusions)
		packages = append(packages, q)
	}

	c.JSON(http.StatusOK, PackagesResponse{
		SessionID: sessionIDStr,
		Packages:  packages,
		TraceID:   traceID,
	})
}
