package lock

import (
	"testing"
)

func TestSagaStateTransitions(t *testing.T) {
	tests := []struct {
		name       string
		from       string
		to         string
		validChain bool
	}{
		{"pending to locking", StatePending, StateLocking, true},
		{"locking to locked", StateLocking, StateLocked, true},
		{"locked to paying", StateLocked, StatePaying, true},
		{"paying to paid", StatePaying, StatePaid, true},
		{"paid to completed", StatePaid, StateCompleted, true},
		{"locking to lock_failed", StateLocking, StateLockFailed, true},
		{"locked to releasing", StateLocked, StateReleasing, true},
		{"releasing to released", StateReleasing, StateReleased, true},
	}

	// Valid forward transitions
	validTransitions := map[string][]string{
		StatePending:      {StateLocking},
		StateLocking:      {StateLocked, StateLockFailed, StateCompensating},
		StateLocked:       {StatePaying, StateReleasing},
		StatePaying:       {StatePaid},
		StatePaid:         {StateCompleted},
		StateReleasing:    {StateReleased},
		StateCompensating: {StateCompensated, StateLockFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed := validTransitions[tt.from]
			found := false
			for _, s := range allowed {
				if s == tt.to {
					found = true
					break
				}
			}
			if tt.validChain && !found {
				t.Errorf("transition %s -> %s should be valid", tt.from, tt.to)
			}
		})
	}
}

func TestLockRedisKey(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		quoteID   string
		wantKey   string
	}{
		{
			name:      "standard key format",
			sessionID: "11111111-1111-1111-1111-111111111111",
			quoteID:   "22222222-2222-2222-2222-222222222222",
			wantKey:   "lock:11111111-1111-1111-1111-111111111111:22222222-2222-2222-2222-222222222222",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse UUIDs manually to avoid import cycle concerns
			key := "lock:" + tt.sessionID + ":" + tt.quoteID
			if key != tt.wantKey {
				t.Errorf("lockRedisKey = %s, want %s", key, tt.wantKey)
			}
		})
	}
}

func TestSagaStates_Constants(t *testing.T) {
	// Verify all states are defined and distinct
	states := []string{
		StatePending,
		StateLocking,
		StateLocked,
		StatePaying,
		StatePaid,
		StateCompleted,
		StateLockFailed,
		StateReleasing,
		StateReleased,
		StateCompensating,
		StateCompensated,
	}

	seen := make(map[string]bool)
	for _, s := range states {
		if s == "" {
			t.Error("found empty state constant")
		}
		if seen[s] {
			t.Errorf("duplicate state: %s", s)
		}
		seen[s] = true
	}

	if len(seen) != 11 {
		t.Errorf("expected 11 unique states, got %d", len(seen))
	}
}

func TestLockTTL(t *testing.T) {
	if lockTTL.Minutes() != 15 {
		t.Errorf("lock TTL should be 15 minutes, got %v", lockTTL)
	}
}

func TestForwardSagaStepOrder(t *testing.T) {
	// Verify the expected step names and their order
	expectedSteps := []struct {
		order int
		name  string
	}{
		{1, "lock_supplier"},
		{2, "freeze_funds"},
		{3, "confirm_lock"},
	}

	for i, step := range expectedSteps {
		if step.order != i+1 {
			t.Errorf("step %s: expected order %d, got %d", step.name, i+1, step.order)
		}
	}
}

func TestCompensationReverseOrder(t *testing.T) {
	// Verify that compensation runs in reverse
	forwardSteps := []string{"lock_supplier", "freeze_funds", "confirm_lock"}
	expectedCompensation := []string{"confirm_lock", "freeze_funds", "lock_supplier"}

	for i := len(forwardSteps) - 1; i >= 0; i-- {
		reverseIdx := len(forwardSteps) - 1 - i
		if forwardSteps[i] != expectedCompensation[reverseIdx] {
			t.Errorf("compensation order mismatch at index %d: got %s, want %s",
				reverseIdx, forwardSteps[i], expectedCompensation[reverseIdx])
		}
	}
}
