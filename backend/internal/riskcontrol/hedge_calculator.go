package riskcontrol

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// SupplierRecoveryRate is the mock supplier refund policy: 80% of base price recoverable.
const SupplierRecoveryRate = 0.80

// HedgeResult holds the calculated refund split between supplier and fund pool.
type HedgeResult struct {
	TotalRefundCents         int64
	SupplierRecoverableCents int64
	PoolCompensationCents    int64
}

// HedgeCalculator computes how a refund is split between supplier recovery and fund pool compensation.
type HedgeCalculator struct {
	db *sql.DB
}

// NewHedgeCalculator creates a new HedgeCalculator.
func NewHedgeCalculator(db *sql.DB) *HedgeCalculator {
	return &HedgeCalculator{db: db}
}

// CalculateRefund computes the hedge split for a given order.
// The user always receives a full refund (total_amount_cents).
// The platform recovers 80% of base_price from the supplier; the fund pool covers the rest.
func (hc *HedgeCalculator) CalculateRefund(ctx context.Context, orderID uuid.UUID, traceID string) (HedgeResult, error) {
	var totalAmountCents, basePriceCents int64
	err := hc.db.QueryRowContext(ctx, `
		SELECT total_amount_cents, base_price_cents
		FROM orders WHERE id = $1`, orderID,
	).Scan(&totalAmountCents, &basePriceCents)
	if err != nil {
		return HedgeResult{}, fmt.Errorf("query order for hedge calculation [trace=%s]: %w", traceID, err)
	}

	supplierRecoverable := int64(float64(basePriceCents) * SupplierRecoveryRate)
	poolCompensation := totalAmountCents - supplierRecoverable

	return HedgeResult{
		TotalRefundCents:         totalAmountCents,
		SupplierRecoverableCents: supplierRecoverable,
		PoolCompensationCents:    poolCompensation,
	}, nil
}

// CalculateRefundFromAmounts is a pure function for testing hedge math without DB.
func CalculateRefundFromAmounts(totalAmountCents, basePriceCents int64) HedgeResult {
	supplierRecoverable := int64(float64(basePriceCents) * SupplierRecoveryRate)
	poolCompensation := totalAmountCents - supplierRecoverable

	return HedgeResult{
		TotalRefundCents:         totalAmountCents,
		SupplierRecoverableCents: supplierRecoverable,
		PoolCompensationCents:    poolCompensation,
	}
}
