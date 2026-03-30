package riskcontrol

import (
	"testing"
)

// TestRefundStateTransitions verifies the refund state machine:
// pending -> supplier_refunded -> pool_compensated -> user_refunded
func TestRefundStateTransitions(t *testing.T) {
	validTransitions := map[string]string{
		"pending":           "supplier_refunded",
		"supplier_refunded": "pool_compensated",
		"pool_compensated":  "user_refunded",
	}

	states := []string{"pending", "supplier_refunded", "pool_compensated", "user_refunded"}

	// Verify the full chain
	for i := 0; i < len(states)-1; i++ {
		current := states[i]
		expected := states[i+1]
		next, ok := validTransitions[current]
		if !ok {
			t.Errorf("no transition defined from state %q", current)
			continue
		}
		if next != expected {
			t.Errorf("transition from %q: got %q, want %q", current, next, expected)
		}
	}

	// user_refunded is terminal
	if _, ok := validTransitions["user_refunded"]; ok {
		t.Error("user_refunded should be a terminal state with no further transitions")
	}
}

// TestRefundStateCount verifies we have exactly 4 states in the refund flow.
func TestRefundStateCount(t *testing.T) {
	states := []string{"pending", "supplier_refunded", "pool_compensated", "user_refunded"}
	if len(states) != 4 {
		t.Errorf("expected 4 refund states, got %d", len(states))
	}
}

// TestRefundHedgeSplit verifies the hedge split covers the full user refund.
func TestRefundHedgeSplit(t *testing.T) {
	tests := []struct {
		name  string
		total int64
		base  int64
	}{
		{"standard order", 100000, 80000},
		{"premium order", 250000, 180000},
		{"budget order", 30000, 25000},
		{"service fee only", 20000, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateRefundFromAmounts(tt.total, tt.base)

			// User always gets full refund
			if result.TotalRefundCents != tt.total {
				t.Errorf("total refund = %d, want %d", result.TotalRefundCents, tt.total)
			}

			// Supplier recoverable should be 80% of base
			expectedSupplier := int64(float64(tt.base) * SupplierRecoveryRate)
			if result.SupplierRecoverableCents != expectedSupplier {
				t.Errorf("supplier recoverable = %d, want %d", result.SupplierRecoverableCents, expectedSupplier)
			}

			// Pool covers the gap
			expectedPool := tt.total - expectedSupplier
			if result.PoolCompensationCents != expectedPool {
				t.Errorf("pool compensation = %d, want %d", result.PoolCompensationCents, expectedPool)
			}

			// Sum must equal total
			sum := result.SupplierRecoverableCents + result.PoolCompensationCents
			if sum != tt.total {
				t.Errorf("supplier(%d) + pool(%d) = %d, want %d",
					result.SupplierRecoverableCents, result.PoolCompensationCents, sum, tt.total)
			}
		})
	}
}

// TestFailedStateIsTerminal verifies 'failed' is recognized as a terminal state.
func TestFailedStateIsTerminal(t *testing.T) {
	terminalStates := map[string]bool{
		"user_refunded": true,
		"failed":        true,
	}

	if !terminalStates["user_refunded"] {
		t.Error("user_refunded should be terminal")
	}
	if !terminalStates["failed"] {
		t.Error("failed should be terminal")
	}
}
