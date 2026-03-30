package payment

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xiaozhang/crayfish-travel/backend/internal/common/middleware"
)

// Handler handles payment-related HTTP requests.
type Handler struct {
	db        *sql.DB
	alipay    AlipayClient
	processor *CallbackProcessor
}

// NewHandler creates a new payment handler.
func NewHandler(db *sql.DB, alipay AlipayClient, processor *CallbackProcessor) *Handler {
	return &Handler{db: db, alipay: alipay, processor: processor}
}

// CreateRequest is the request body for creating a payment.
type CreateRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	QuoteID   string `json:"quote_id" binding:"required"`
	Method    string `json:"method" binding:"required,oneof=qr voice_token"`
}

// CreateResponse contains the payment creation result.
type CreateResponse struct {
	PaymentID  string `json:"payment_id"`
	OutTradeNo string `json:"out_trade_no"`
	QRCodeURL  string `json:"qr_code_url,omitempty"`
	VoiceToken string `json:"voice_token,omitempty"`
	Method     string `json:"method"`
	TraceID    string `json:"trace_id"`
}

// Create godoc
// @Summary      Create payment
// @Description  Generate QR code or voice token for Alipay payment
// @Tags         payment
// @Accept       json
// @Produce      json
// @Param        body  body  CreateRequest  true  "Payment details"
// @Success      201   {object}  CreateResponse
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /payment/create [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
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

	// Get quote total price
	var totalPriceCents int64
	var packageTitle string
	err = h.db.QueryRow(`
		SELECT total_price_cents, package_title FROM supplier_quotes
		WHERE id = $1 AND session_id = $2`,
		quoteID, sessionID,
	).Scan(&totalPriceCents, &packageTitle)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "quote not found"})
		return
	}

	outTradeNo := GenerateOutTradeNo()

	// Generate payment via Alipay
	var result *PaymentResult
	switch req.Method {
	case "qr":
		result, err = h.alipay.CreateQRPayment(outTradeNo, totalPriceCents, packageTitle)
	case "voice_token":
		result, err = h.alipay.CreateVoiceToken(outTradeNo, totalPriceCents, packageTitle)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create payment"})
		return
	}

	// Save payment record
	var paymentID uuid.UUID
	err = h.db.QueryRow(`
		INSERT INTO payments (session_id, lock_session_id, quote_id, trace_id, out_trade_no, method, amount_cents, qr_code_url, voice_token)
		VALUES ($1, (SELECT id FROM lock_sessions WHERE session_id = $1 AND quote_id = $2 LIMIT 1), $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`,
		sessionID, quoteID, traceID, outTradeNo, req.Method, totalPriceCents,
		result.QRCodeURL, result.VoiceToken,
	).Scan(&paymentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save payment"})
		return
	}

	c.JSON(http.StatusCreated, CreateResponse{
		PaymentID:  paymentID.String(),
		OutTradeNo: outTradeNo,
		QRCodeURL:  result.QRCodeURL,
		VoiceToken: result.VoiceToken,
		Method:     result.Method,
		TraceID:    traceID,
	})
}

// Callback godoc
// @Summary      Alipay payment callback
// @Description  Handle Alipay payment notification (signature verified, idempotent)
// @Tags         payment
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Router       /payment/callback [post]
func (h *Handler) Callback(c *gin.Context) {
	traceID := middleware.GetTraceID(c)

	var params CallbackParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify signature
	rawParams := map[string]string{
		"out_trade_no": params.OutTradeNo,
		"trade_no":     params.TradeNo,
		"trade_status": params.TradeStatus,
		"total_amount": params.TotalAmount,
	}
	valid, err := h.alipay.VerifyCallback(rawParams)
	if err != nil || !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature"})
		return
	}

	if err := h.processor.ProcessCallback(c.Request.Context(), params, traceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "callback processing failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
