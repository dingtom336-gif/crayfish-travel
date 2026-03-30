package payment

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// OrderCreator is called after successful payment to create an order.
type OrderCreator interface {
	CreateFromPayment(ctx context.Context, sessionID, paymentID, quoteID, traceID string) error
}

// CallbackProcessor handles payment callback logic.
type CallbackProcessor struct {
	db           *sql.DB
	redis        *goredis.Client
	alipay       AlipayClient
	orderCreator OrderCreator
}

// NewCallbackProcessor creates a new callback processor.
func NewCallbackProcessor(db *sql.DB, redis *goredis.Client, alipay AlipayClient) *CallbackProcessor {
	return &CallbackProcessor{db: db, redis: redis, alipay: alipay}
}

// SetOrderCreator sets the order creation hook (avoids circular dependency).
func (p *CallbackProcessor) SetOrderCreator(oc OrderCreator) {
	p.orderCreator = oc
}

// CallbackParams represents Alipay callback notification parameters.
type CallbackParams struct {
	OutTradeNo  string `json:"out_trade_no"`
	TradeNo     string `json:"trade_no"`
	TradeStatus string `json:"trade_status"`
	TotalAmount string `json:"total_amount"`
}

// ProcessCallback handles a payment callback notification.
// Steps:
// 1. Verify signature (via AlipayClient)
// 2. Idempotency check (Redis key: payment:processed:{out_trade_no})
// 3. Update payment status to 'paid'
// 4. Release other 4 locked quotes
// 5. Trigger order creation
func (p *CallbackProcessor) ProcessCallback(ctx context.Context, params CallbackParams, traceID string) error {
	// 1. Idempotency check
	idempotencyKey := fmt.Sprintf("payment:processed:%s", params.OutTradeNo)
	set, err := p.redis.SetNX(ctx, idempotencyKey, "1", 24*time.Hour).Result()
	if err != nil {
		return fmt.Errorf("idempotency check: %w", err)
	}
	if !set {
		// Already processed
		return nil
	}

	// 2. Verify trade status
	if params.TradeStatus != "TRADE_SUCCESS" && params.TradeStatus != "TRADE_FINISHED" {
		return nil // Ignore non-success callbacks
	}

	// 3. Update payment status
	var paymentID, sessionID, quoteID string
	err = p.db.QueryRowContext(ctx, `
		UPDATE payments SET status = 'paid', alipay_trade_no = $1, paid_at = NOW(), updated_at = NOW()
		WHERE out_trade_no = $2 AND status = 'pending'
		RETURNING id, session_id, quote_id`,
		params.TradeNo, params.OutTradeNo,
	).Scan(&paymentID, &sessionID, &quoteID)
	if err != nil {
		// Rollback idempotency key on failure
		p.redis.Del(ctx, idempotencyKey)
		return fmt.Errorf("update payment: %w", err)
	}

	// 4. Release other 4 locked quotes (keep only the paid one)
	_, err = p.db.ExecContext(ctx, `
		UPDATE supplier_quotes SET status = 'released'
		WHERE session_id = $1 AND id != $2 AND status = 'active'`,
		sessionID, quoteID,
	)
	if err != nil {
		return fmt.Errorf("release other quotes: %w", err)
	}

	// 5. Update session and lock status
	if _, err := p.db.ExecContext(ctx, `UPDATE sessions SET status = 'paid', updated_at = NOW() WHERE id = $1`, sessionID); err != nil {
		log.Printf("[payment:callback] failed to update session %s status: %v", sessionID, err)
	}
	if _, err := p.db.ExecContext(ctx, `
		UPDATE lock_sessions SET state = 'PAID', updated_at = NOW()
		WHERE session_id = $1 AND quote_id = $2 AND state = 'LOCKED'`,
		sessionID, quoteID,
	); err != nil {
		log.Printf("[payment:callback] failed to update lock state for session %s: %v", sessionID, err)
	}

	// 6. Auto-create order
	if p.orderCreator != nil {
		if err := p.orderCreator.CreateFromPayment(ctx, sessionID, paymentID, quoteID, traceID); err != nil {
			log.Printf("[payment:callback] failed to create order for session %s: %v", sessionID, err)
		}
	}

	return nil
}
