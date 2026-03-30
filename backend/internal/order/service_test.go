package order

import (
	"strings"
	"testing"
)

func TestGenerateOrderNo(t *testing.T) {
	orderNo := GenerateOrderNo()

	if !strings.HasPrefix(orderNo, "CO") {
		t.Errorf("order number should start with 'CO', got %q", orderNo)
	}

	// "CO" (2) + timestamp "20060102150405" (14) + random (8) = 24
	expectedLen := 24
	if len(orderNo) != expectedLen {
		t.Errorf("order number length should be %d, got %d (%q)", expectedLen, len(orderNo), orderNo)
	}

	// Suffix (last 8 chars) should be alphanumeric uppercase
	suffix := orderNo[16:]
	for _, ch := range suffix {
		if !((ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')) {
			t.Errorf("suffix should be alphanumeric uppercase, got char %q in %q", string(ch), suffix)
		}
	}

	// Uniqueness: generate multiple and check no duplicates
	seen := map[string]bool{orderNo: true}
	for i := 0; i < 100; i++ {
		no := GenerateOrderNo()
		if seen[no] {
			t.Errorf("duplicate order number generated: %q", no)
		}
		seen[no] = true
	}
}

func TestOrderStatusTransitions(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus string
		canRefund     bool
	}{
		{"created allows refund", "created", true},
		{"confirmed allows refund", "confirmed", true},
		{"fulfilling allows refund", "fulfilling", true},
		{"completed allows refund", "completed", true},
		{"refund_requested blocks refund", "refund_requested", false},
		{"refunded blocks refund", "refunded", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := refundableStatuses[tt.currentStatus]
			if got != tt.canRefund {
				t.Errorf("status %q: refundable = %v, want %v", tt.currentStatus, got, tt.canRefund)
			}
		})
	}
}
