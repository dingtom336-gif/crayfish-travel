package riskcontrol

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiaozhang/crayfish-travel/backend/internal/common/middleware"
)

// Handler handles risk-control HTTP requests.
type Handler struct {
	pool      *FundPool
	antifraud *AntifraudChecker
}

// NewHandler creates a new risk-control handler.
func NewHandler(pool *FundPool, antifraud *AntifraudChecker) *Handler {
	return &Handler{pool: pool, antifraud: antifraud}
}

// PoolStatusResponse contains the fund pool balance.
type PoolStatusResponse struct {
	AvailableCents int64  `json:"available_cents"`
	FrozenCents    int64  `json:"frozen_cents"`
	TotalCents     int64  `json:"total_cents"`
	TraceID        string `json:"trace_id"`
}

// SeedRequest is the request body for seeding the fund pool.
type SeedRequest struct {
	AmountCents int64 `json:"amount_cents" binding:"required,min=1"`
}

// Seed godoc
// @Summary      Seed fund pool (admin, dev only)
// @Description  Add initial balance to the risk-control fund pool
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        body  body  SeedRequest  true  "Seed amount"
// @Success      200  {object}  PoolStatusResponse
// @Failure      400  {object}  map[string]string
// @Router       /admin/fund-pool/seed [post]
func (h *Handler) Seed(c *gin.Context) {
	var req SeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	traceID := middleware.GetTraceID(c)

	if err := h.pool.Deposit(req.AmountCents, traceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	balance, _ := h.pool.GetBalance()
	c.JSON(http.StatusOK, PoolStatusResponse{
		AvailableCents: balance.AvailableCents,
		FrozenCents:    balance.FrozenCents,
		TotalCents:     balance.TotalCents,
		TraceID:        traceID,
	})
}

// PoolStatus godoc
// @Summary      Get fund pool status
// @Description  Returns available, frozen, and total balance of the risk-control fund pool
// @Tags         riskcontrol
// @Produce      json
// @Success      200  {object}  PoolStatusResponse
// @Failure      500  {object}  map[string]string
// @Router       /riskcontrol/pool/status [get]
func (h *Handler) PoolStatus(c *gin.Context) {
	traceID := middleware.GetTraceID(c)

	balance, err := h.pool.GetBalance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get pool status", "trace_id": traceID})
		return
	}

	c.JSON(http.StatusOK, PoolStatusResponse{
		AvailableCents: balance.AvailableCents,
		FrozenCents:    balance.FrozenCents,
		TotalCents:     balance.TotalCents,
		TraceID:        traceID,
	})
}
