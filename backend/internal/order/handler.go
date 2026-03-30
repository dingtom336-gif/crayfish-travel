package order

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xiaozhang/crayfish-travel/backend/internal/common/middleware"
)

// RefundProcessorInterface defines the interface for processing refunds.
type RefundProcessorInterface interface {
	ProcessRefund(ctx context.Context, orderID uuid.UUID, traceID string) error
}

// Handler handles order-related HTTP requests.
type Handler struct {
	service         *OrderService
	notifier        SmsNotifier
	refundProcessor RefundProcessorInterface
}

// NewHandler creates a new order handler.
func NewHandler(service *OrderService, notifier SmsNotifier, refundProcessor RefundProcessorInterface) *Handler {
	return &Handler{service: service, notifier: notifier, refundProcessor: refundProcessor}
}

// OrderListResponse wraps a list of orders.
type OrderListResponse struct {
	Orders []Order `json:"orders"`
}

// List godoc
// @Summary      List orders
// @Description  List all orders for a given session
// @Tags         order
// @Produce      json
// @Param        session_id  query  string  true  "Session ID"
// @Success      200  {object}  OrderListResponse
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /orders [get]
func (h *Handler) List(c *gin.Context) {
	sessionIDStr := c.Query("session_id")
	if sessionIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session_id"})
		return
	}

	orders, err := h.service.ListBySession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list orders"})
		return
	}

	if orders == nil {
		orders = []Order{}
	}

	c.JSON(http.StatusOK, OrderListResponse{Orders: orders})
}

// Get godoc
// @Summary      Get order
// @Description  Get a single order with full details
// @Tags         order
// @Produce      json
// @Param        id  path  string  true  "Order ID"
// @Success      200  {object}  Order
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /orders/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	order, err := h.service.GetByID(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

// RefundRequest is the response for a refund request.
type RefundRequest struct {
	Status  string `json:"status"`
	TraceID string `json:"trace_id"`
}

// RequestRefund godoc
// @Summary      Request refund
// @Description  Request a refund for an order (updates status to refund_requested)
// @Tags         order
// @Produce      json
// @Param        id  path  string  true  "Order ID"
// @Success      200  {object}  RefundRequest
// @Failure      400  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Router       /orders/{id}/refund [post]
func (h *Handler) RequestRefund(c *gin.Context) {
	traceID := middleware.GetTraceID(c)

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	if err := h.service.RequestRefund(c.Request.Context(), orderID, traceID); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// Process the full refund flow asynchronously
	if h.refundProcessor != nil {
		go func() {
			bgCtx := context.Background()
			if err := h.refundProcessor.ProcessRefund(bgCtx, orderID, traceID); err != nil {
				log.Printf("[OrderHandler] refund processing failed for order=%s trace=%s: %v", orderID, traceID, err)
			}
		}()
	}

	c.JSON(http.StatusOK, RefundRequest{
		Status:  "refund_requested",
		TraceID: traceID,
	})
}
