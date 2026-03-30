package riskcontrol

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// RefundProcessor orchestrates the full refund flow:
// anti-fraud check -> hedge calculation -> supplier refund -> pool compensation -> user refund.
type RefundProcessor struct {
	db              *sql.DB
	fundPool        *FundPool
	hedgeCalculator *HedgeCalculator
	antifraud       *AntifraudChecker
}

// NewRefundProcessor creates a new RefundProcessor.
func NewRefundProcessor(db *sql.DB, fundPool *FundPool, hedgeCalculator *HedgeCalculator, antifraud *AntifraudChecker) *RefundProcessor {
	return &RefundProcessor{
		db:              db,
		fundPool:        fundPool,
		hedgeCalculator: hedgeCalculator,
		antifraud:       antifraud,
	}
}

// ProcessRefund executes the complete refund state machine for an order.
func (rp *RefundProcessor) ProcessRefund(ctx context.Context, orderID uuid.UUID, traceID string) error {
	log.Printf("[RefundProcessor] starting refund for order=%s trace=%s", orderID, traceID)

	// Step 1: Get order info for anti-fraud check
	var sessionID uuid.UUID
	var totalAmountCents int64
	err := rp.db.QueryRowContext(ctx, `
		SELECT session_id, total_amount_cents FROM orders WHERE id = $1`, orderID,
	).Scan(&sessionID, &totalAmountCents)
	if err != nil {
		return fmt.Errorf("query order for refund [trace=%s]: %w", traceID, err)
	}

	// Get identity hash from session for anti-fraud check
	var idHash string
	err = rp.db.QueryRowContext(ctx, `
		SELECT id_hash FROM identity_records
		WHERE session_id = $1 ORDER BY created_at DESC LIMIT 1`, sessionID,
	).Scan(&idHash)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("query identity for anti-fraud [trace=%s]: %w", traceID, err)
	}

	// Step 2: Anti-fraud check
	if idHash != "" {
		if err := rp.antifraud.CheckRefundFrequency(ctx, idHash); err != nil {
			return fmt.Errorf("anti-fraud check failed [trace=%s]: %w", traceID, err)
		}
	}

	// Step 3: Calculate hedge amounts
	hedge, err := rp.hedgeCalculator.CalculateRefund(ctx, orderID, traceID)
	if err != nil {
		return fmt.Errorf("hedge calculation failed [trace=%s]: %w", traceID, err)
	}

	log.Printf("[RefundProcessor] hedge result: total=%d supplier=%d pool=%d trace=%s",
		hedge.TotalRefundCents, hedge.SupplierRecoverableCents, hedge.PoolCompensationCents, traceID)

	// Step 4: Create refund request record (status: pending)
	refundID, err := rp.createRefundRequest(ctx, orderID, sessionID, traceID, hedge)
	if err != nil {
		return fmt.Errorf("create refund request [trace=%s]: %w", traceID, err)
	}

	// Step 5: Simulate supplier refund (mock: always succeeds)
	log.Printf("[RefundProcessor] simulating supplier refund of %d cents trace=%s", hedge.SupplierRecoverableCents, traceID)
	if err := rp.updateRefundStatus(ctx, refundID, "supplier_refunded"); err != nil {
		return fmt.Errorf("update to supplier_refunded [trace=%s]: %w", traceID, err)
	}
	if _, err := rp.db.ExecContext(ctx, `
		UPDATE refund_requests SET supplier_refund_at = $1, updated_at = $1 WHERE id = $2`,
		time.Now(), refundID,
	); err != nil {
		return fmt.Errorf("set supplier_refund_at [trace=%s]: %w", traceID, err)
	}

	// Step 6: Fund pool compensation via REFUND ledger entry
	if hedge.PoolCompensationCents > 0 {
		if _, err := rp.db.ExecContext(ctx, `
			INSERT INTO fund_pool_ledger (trace_id, session_id, operation, amount_cents, balance_after_cents, description)
			VALUES ($1, $2, 'REFUND', $3, 0, 'Pool compensation for refund')`,
			traceID, sessionID, hedge.PoolCompensationCents,
		); err != nil {
			return fmt.Errorf("fund pool compensation [trace=%s]: %w", traceID, err)
		}
	}
	if err := rp.updateRefundStatus(ctx, refundID, "pool_compensated"); err != nil {
		return fmt.Errorf("update to pool_compensated [trace=%s]: %w", traceID, err)
	}

	// Step 7: Mark user refund complete
	if err := rp.updateRefundStatus(ctx, refundID, "user_refunded"); err != nil {
		return fmt.Errorf("update to user_refunded [trace=%s]: %w", traceID, err)
	}
	if _, err := rp.db.ExecContext(ctx, `
		UPDATE refund_requests SET user_refund_at = $1, updated_at = $1 WHERE id = $2`,
		time.Now(), refundID,
	); err != nil {
		return fmt.Errorf("set user_refund_at [trace=%s]: %w", traceID, err)
	}

	// Step 8: Update order status to 'refunded'
	if _, err := rp.db.ExecContext(ctx, `
		UPDATE orders SET status = 'refunded', updated_at = NOW() WHERE id = $1`, orderID,
	); err != nil {
		return fmt.Errorf("update order to refunded [trace=%s]: %w", traceID, err)
	}

	// Step 9: Increment anti-fraud refund counter
	if idHash != "" {
		if err := rp.antifraud.IncrementRefund(ctx, idHash); err != nil {
			log.Printf("[RefundProcessor] WARNING: failed to increment refund counter [trace=%s]: %v", traceID, err)
		}
	}

	// Step 10: Update session status to 'refunded'
	if _, err := rp.db.ExecContext(ctx, `
		UPDATE sessions SET status = 'refunded', updated_at = NOW() WHERE id = $1`, sessionID,
	); err != nil {
		return fmt.Errorf("update session to refunded [trace=%s]: %w", traceID, err)
	}

	log.Printf("[RefundProcessor] refund completed for order=%s refund=%s trace=%s", orderID, refundID, traceID)
	return nil
}

// createRefundRequest inserts a new refund_requests record with status 'pending'.
func (rp *RefundProcessor) createRefundRequest(ctx context.Context, orderID, sessionID uuid.UUID, traceID string, hedge HedgeResult) (uuid.UUID, error) {
	var refundID uuid.UUID
	err := rp.db.QueryRowContext(ctx, `
		INSERT INTO refund_requests (order_id, session_id, trace_id, status,
			total_refund_cents, supplier_recoverable_cents, pool_compensation_cents)
		VALUES ($1, $2, $3, 'pending', $4, $5, $6)
		RETURNING id`,
		orderID, sessionID, traceID,
		hedge.TotalRefundCents, hedge.SupplierRecoverableCents, hedge.PoolCompensationCents,
	).Scan(&refundID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert refund request: %w", err)
	}
	return refundID, nil
}

// updateRefundStatus transitions the refund request to a new status.
func (rp *RefundProcessor) updateRefundStatus(ctx context.Context, refundID uuid.UUID, status string) error {
	_, err := rp.db.ExecContext(ctx, `
		UPDATE refund_requests SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, refundID,
	)
	if err != nil {
		return fmt.Errorf("update refund status to %s: %w", status, err)
	}
	return nil
}
