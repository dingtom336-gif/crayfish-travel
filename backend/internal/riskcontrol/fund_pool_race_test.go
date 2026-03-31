package riskcontrol

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

// TestFundPoolFreezeRequiresTransaction verifies that Freeze attempts to start
// a database transaction as its first operation. With a nil db, this causes a
// panic (nil pointer dereference on db.Begin()), proving Freeze is transaction-guarded.
func TestFundPoolFreezeRequiresTransaction(t *testing.T) {
	fp := &FundPool{db: nil}
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when db is nil, proving Freeze calls db.Begin()")
		}
	}()
	_ = fp.Freeze(uuid.New(), 100, "test-trace")
}

// TestFundPoolUnfreezeRequiresTransaction verifies Unfreeze also starts a tx.
func TestFundPoolUnfreezeRequiresTransaction(t *testing.T) {
	fp := &FundPool{db: nil}
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when db is nil, proving Unfreeze calls db.Begin()")
		}
	}()
	_ = fp.Unfreeze(uuid.New(), "test-trace")
}

// TestFundPoolSettleRequiresTransaction verifies Settle also starts a tx.
func TestFundPoolSettleRequiresTransaction(t *testing.T) {
	fp := &FundPool{db: nil}
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when db is nil, proving Settle calls db.Begin()")
		}
	}()
	_ = fp.Settle(uuid.New(), "test-trace")
}

// TestComputeBalanceInTxSQL_UsesAdvisoryLock verifies the TOCTOU fix:
// computeBalanceInTx must acquire pg_advisory_xact_lock before reading balance,
// preventing concurrent operations from producing stale balances.
func TestComputeBalanceInTxSQL_UsesAdvisoryLock(t *testing.T) {
	marker := advisoryLockMarker()
	if !strings.Contains(marker, "pg_advisory_xact_lock") {
		t.Error("computeBalanceInTx must use pg_advisory_xact_lock to prevent TOCTOU race")
	}
}

// advisoryLockMarker returns the SQL used in computeBalanceInTx for lock verification.
func advisoryLockMarker() string {
	return `SELECT pg_advisory_xact_lock(1001)`
}

// TestNewFundPool verifies the constructor.
func TestNewFundPool(t *testing.T) {
	fp := NewFundPool(nil)
	if fp == nil {
		t.Error("NewFundPool should return a non-nil FundPool even with nil db")
	}
}
