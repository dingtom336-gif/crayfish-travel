package riskcontrol

import (
	"testing"
)

// TestHedgeCalculation verifies 80% supplier recovery and correct pool compensation math.
func TestHedgeCalculation(t *testing.T) {
	tests := []struct {
		name                     string
		totalAmountCents         int64
		basePriceCents           int64
		wantSupplierRecoverable  int64
		wantPoolCompensation     int64
		wantTotalRefund          int64
	}{
		{
			name:                     "standard booking",
			totalAmountCents:         100000, // 1000.00
			basePriceCents:           80000,  // 800.00
			wantSupplierRecoverable:  64000,  // 80% of 800.00
			wantPoolCompensation:     36000,  // 1000.00 - 640.00
			wantTotalRefund:          100000,
		},
		{
			name:                     "high service fee booking",
			totalAmountCents:         150000, // 1500.00
			basePriceCents:           100000, // 1000.00
			wantSupplierRecoverable:  80000,  // 80% of 1000.00
			wantPoolCompensation:     70000,  // 1500.00 - 800.00
			wantTotalRefund:          150000,
		},
		{
			name:                     "small booking",
			totalAmountCents:         10000, // 100.00
			basePriceCents:           8000,  // 80.00
			wantSupplierRecoverable:  6400,  // 80% of 80.00
			wantPoolCompensation:     3600,  // 100.00 - 64.00
			wantTotalRefund:          10000,
		},
		{
			name:                     "zero base price",
			totalAmountCents:         5000, // 50.00 (all service fee)
			basePriceCents:           0,
			wantSupplierRecoverable:  0,
			wantPoolCompensation:     5000, // pool covers everything
			wantTotalRefund:          5000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateRefundFromAmounts(tt.totalAmountCents, tt.basePriceCents)

			if result.TotalRefundCents != tt.wantTotalRefund {
				t.Errorf("TotalRefundCents = %d, want %d", result.TotalRefundCents, tt.wantTotalRefund)
			}
			if result.SupplierRecoverableCents != tt.wantSupplierRecoverable {
				t.Errorf("SupplierRecoverableCents = %d, want %d", result.SupplierRecoverableCents, tt.wantSupplierRecoverable)
			}
			if result.PoolCompensationCents != tt.wantPoolCompensation {
				t.Errorf("PoolCompensationCents = %d, want %d", result.PoolCompensationCents, tt.wantPoolCompensation)
			}
		})
	}
}

// TestFullRefundToUser verifies the user always gets total_amount_cents back.
func TestFullRefundToUser(t *testing.T) {
	testCases := []struct {
		total int64
		base  int64
	}{
		{100000, 80000},
		{50000, 40000},
		{200000, 150000},
		{10000, 0},
		{999999, 750000},
	}

	for _, tc := range testCases {
		result := CalculateRefundFromAmounts(tc.total, tc.base)
		if result.TotalRefundCents != tc.total {
			t.Errorf("user refund for total=%d: got %d, want %d",
				tc.total, result.TotalRefundCents, tc.total)
		}

		// Verify: supplier + pool = total
		if result.SupplierRecoverableCents+result.PoolCompensationCents != tc.total {
			t.Errorf("supplier(%d) + pool(%d) = %d, want %d",
				result.SupplierRecoverableCents, result.PoolCompensationCents,
				result.SupplierRecoverableCents+result.PoolCompensationCents, tc.total)
		}
	}
}

// TestSupplierRecoveryRate verifies the 80% recovery rate constant.
func TestSupplierRecoveryRate(t *testing.T) {
	if SupplierRecoveryRate != 0.80 {
		t.Errorf("SupplierRecoveryRate = %f, want 0.80", SupplierRecoveryRate)
	}
}
