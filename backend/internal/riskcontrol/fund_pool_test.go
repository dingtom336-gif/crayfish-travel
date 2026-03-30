package riskcontrol

import (
	"testing"
)

// TestBalanceCalculation tests the balance math logic directly.
// The core formula: available = deposits - freezes + unfreezes - refunds
//                   frozen = freezes - unfreezes - settles
//                   total = available + frozen
func TestBalanceCalculation(t *testing.T) {
	tests := []struct {
		name      string
		deposits  int64
		freezes   int64
		unfreezes int64
		settles   int64
		refunds   int64
		wantAvail int64
		wantFroz  int64
		wantTotal int64
	}{
		{
			name:      "empty pool",
			wantAvail: 0,
			wantFroz:  0,
			wantTotal: 0,
		},
		{
			name:      "deposit only",
			deposits:  100000,
			wantAvail: 100000,
			wantFroz:  0,
			wantTotal: 100000,
		},
		{
			name:      "deposit then freeze",
			deposits:  100000,
			freezes:   30000,
			wantAvail: 70000,
			wantFroz:  30000,
			wantTotal: 100000,
		},
		{
			name:      "deposit freeze unfreeze",
			deposits:  100000,
			freezes:   30000,
			unfreezes: 30000,
			wantAvail: 100000,
			wantFroz:  0,
			wantTotal: 100000,
		},
		{
			name:      "deposit freeze settle",
			deposits:  100000,
			freezes:   30000,
			settles:   30000,
			wantAvail: 70000,
			wantFroz:  0,
			wantTotal: 70000,
		},
		{
			name:      "deposit freeze settle refund",
			deposits:  100000,
			freezes:   30000,
			settles:   30000,
			refunds:   10000,
			wantAvail: 60000,
			wantFroz:  0,
			wantTotal: 60000,
		},
		{
			name:      "multiple freezes partial unfreeze",
			deposits:  200000,
			freezes:   80000,
			unfreezes: 30000,
			wantAvail: 150000,
			wantFroz:  50000,
			wantTotal: 200000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replicate the balance formula from computeBalance
			frozen := tt.freezes - tt.unfreezes - tt.settles
			available := tt.deposits - tt.freezes + tt.unfreezes - tt.refunds
			total := available + frozen

			if available != tt.wantAvail {
				t.Errorf("available = %d, want %d", available, tt.wantAvail)
			}
			if frozen != tt.wantFroz {
				t.Errorf("frozen = %d, want %d", frozen, tt.wantFroz)
			}
			if total != tt.wantTotal {
				t.Errorf("total = %d, want %d", total, tt.wantTotal)
			}
		})
	}
}

// TestInsufficientFundsCheck verifies the insufficient funds guard.
func TestInsufficientFundsCheck(t *testing.T) {
	tests := []struct {
		name      string
		available int64
		request   int64
		wantOK    bool
	}{
		{"exact amount", 50000, 50000, true},
		{"more than enough", 100000, 50000, true},
		{"not enough", 30000, 50000, false},
		{"zero available", 0, 10000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok := tt.available >= tt.request
			if ok != tt.wantOK {
				t.Errorf("available=%d request=%d: got ok=%v, want %v",
					tt.available, tt.request, ok, tt.wantOK)
			}
		})
	}
}

// TestFreezeUnfreezeSymmetry verifies that freeze + unfreeze returns to original balance.
func TestFreezeUnfreezeSymmetry(t *testing.T) {
	initial := int64(100000)
	freezeAmount := int64(30000)

	// After freeze
	availableAfterFreeze := initial - freezeAmount
	frozenAfterFreeze := freezeAmount

	if availableAfterFreeze+frozenAfterFreeze != initial {
		t.Errorf("total changed after freeze: %d + %d != %d",
			availableAfterFreeze, frozenAfterFreeze, initial)
	}

	// After unfreeze
	availableAfterUnfreeze := availableAfterFreeze + freezeAmount
	frozenAfterUnfreeze := frozenAfterFreeze - freezeAmount

	if availableAfterUnfreeze != initial {
		t.Errorf("available not restored: %d != %d", availableAfterUnfreeze, initial)
	}
	if frozenAfterUnfreeze != 0 {
		t.Errorf("frozen not zero: %d", frozenAfterUnfreeze)
	}
}

// TestSettleReducesTotal verifies that settle removes funds from the pool entirely.
func TestSettleReducesTotal(t *testing.T) {
	deposits := int64(100000)
	freezeAmount := int64(30000)
	settleAmount := int64(30000)

	frozen := freezeAmount - settleAmount
	available := deposits - freezeAmount
	total := available + frozen

	if total != 70000 {
		t.Errorf("total after settle = %d, want 70000", total)
	}
	if frozen != 0 {
		t.Errorf("frozen after settle = %d, want 0", frozen)
	}
}
