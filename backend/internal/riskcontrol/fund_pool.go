package riskcontrol

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// BalanceInfo holds the current fund pool state.
type BalanceInfo struct {
	AvailableCents int64 `json:"available_cents"`
	FrozenCents    int64 `json:"frozen_cents"`
	TotalCents     int64 `json:"total_cents"`
}

// FundPool manages the risk-control fund pool with atomic ledger operations.
type FundPool struct {
	db *sql.DB
}

// NewFundPool creates a new FundPool.
func NewFundPool(db *sql.DB) *FundPool {
	return &FundPool{db: db}
}

// Freeze atomically freezes funds for a booking session.
func (fp *FundPool) Freeze(sessionID uuid.UUID, amountCents int64, traceID string) error {
	tx, err := fp.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	balance, err := fp.computeBalanceInTx(tx)
	if err != nil {
		return fmt.Errorf("failed to compute balance: %w", err)
	}

	if balance.AvailableCents < amountCents {
		return fmt.Errorf("insufficient funds: available %d, requested %d", balance.AvailableCents, amountCents)
	}

	newBalance := balance.AvailableCents - amountCents
	_, err = tx.Exec(`
		INSERT INTO fund_pool_ledger (trace_id, session_id, operation, amount_cents, balance_after_cents, description)
		VALUES ($1, $2, 'FREEZE', $3, $4, 'Freeze funds for booking')`,
		traceID, sessionID, amountCents, newBalance,
	)
	if err != nil {
		return fmt.Errorf("failed to insert ledger entry: %w", err)
	}

	return tx.Commit()
}

// Deposit adds funds to the pool (admin operation).
func (fp *FundPool) Deposit(amountCents int64, traceID string) error {
	var totalCents int64
	balance, err := fp.GetBalance()
	if err == nil {
		totalCents = balance.TotalCents
	}
	newBalance := totalCents + amountCents

	_, err = fp.db.Exec(`
		INSERT INTO fund_pool_ledger (trace_id, operation, amount_cents, balance_after_cents, description)
		VALUES ($1, 'DEPOSIT', $2, $3, 'Admin seed deposit')`,
		traceID, amountCents, newBalance,
	)
	if err != nil {
		return fmt.Errorf("failed to deposit: %w", err)
	}
	return nil
}

// Unfreeze releases previously frozen funds back to available.
func (fp *FundPool) Unfreeze(sessionID uuid.UUID, traceID string) error {
	tx, err := fp.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	// Find the frozen amount for this session
	var frozenAmount int64
	err = tx.QueryRow(`
		SELECT COALESCE(SUM(CASE WHEN operation = 'FREEZE' THEN amount_cents ELSE 0 END) -
		       SUM(CASE WHEN operation IN ('UNFREEZE', 'SETTLE') THEN amount_cents ELSE 0 END), 0)
		FROM fund_pool_ledger WHERE session_id = $1`,
		sessionID,
	).Scan(&frozenAmount)
	if err != nil {
		return fmt.Errorf("failed to query frozen amount: %w", err)
	}

	if frozenAmount <= 0 {
		return fmt.Errorf("no frozen funds for session %s", sessionID)
	}

	balance, err := fp.computeBalanceInTx(tx)
	if err != nil {
		return fmt.Errorf("failed to compute balance: %w", err)
	}

	newBalance := balance.AvailableCents + frozenAmount
	_, err = tx.Exec(`
		INSERT INTO fund_pool_ledger (trace_id, session_id, operation, amount_cents, balance_after_cents, description)
		VALUES ($1, $2, 'UNFREEZE', $3, $4, 'Release frozen funds')`,
		traceID, sessionID, frozenAmount, newBalance,
	)
	if err != nil {
		return fmt.Errorf("failed to insert ledger entry: %w", err)
	}

	return tx.Commit()
}

// Settle marks frozen funds as settled after successful payment.
func (fp *FundPool) Settle(sessionID uuid.UUID, traceID string) error {
	tx, err := fp.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	// Find the frozen amount for this session
	var frozenAmount int64
	err = tx.QueryRow(`
		SELECT COALESCE(SUM(CASE WHEN operation = 'FREEZE' THEN amount_cents ELSE 0 END) -
		       SUM(CASE WHEN operation IN ('UNFREEZE', 'SETTLE') THEN amount_cents ELSE 0 END), 0)
		FROM fund_pool_ledger WHERE session_id = $1`,
		sessionID,
	).Scan(&frozenAmount)
	if err != nil {
		return fmt.Errorf("failed to query frozen amount: %w", err)
	}

	if frozenAmount <= 0 {
		return fmt.Errorf("no frozen funds to settle for session %s", sessionID)
	}

	balance, err := fp.computeBalanceInTx(tx)
	if err != nil {
		return fmt.Errorf("failed to compute balance: %w", err)
	}

	// Settle removes from frozen (already not in available), balance stays same
	_, err = tx.Exec(`
		INSERT INTO fund_pool_ledger (trace_id, session_id, operation, amount_cents, balance_after_cents, description)
		VALUES ($1, $2, 'SETTLE', $3, $4, 'Settle funds after payment')`,
		traceID, sessionID, frozenAmount, balance.AvailableCents,
	)
	if err != nil {
		return fmt.Errorf("failed to insert ledger entry: %w", err)
	}

	return tx.Commit()
}

// GetBalance returns current available, frozen, and total amounts.
func (fp *FundPool) GetBalance() (BalanceInfo, error) {
	return fp.computeBalance()
}

// computeBalance calculates the current balance from the ledger.
func (fp *FundPool) computeBalance() (BalanceInfo, error) {
	var deposits, freezes, unfreezes, settles, refunds int64
	err := fp.db.QueryRow(`
		SELECT
			COALESCE(SUM(CASE WHEN operation = 'DEPOSIT' THEN amount_cents ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN operation = 'FREEZE' THEN amount_cents ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN operation = 'UNFREEZE' THEN amount_cents ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN operation = 'SETTLE' THEN amount_cents ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN operation = 'REFUND' THEN amount_cents ELSE 0 END), 0)
		FROM fund_pool_ledger`,
	).Scan(&deposits, &freezes, &unfreezes, &settles, &refunds)
	if err != nil {
		return BalanceInfo{}, fmt.Errorf("failed to compute balance: %w", err)
	}

	frozen := freezes - unfreezes - settles
	available := deposits - freezes + unfreezes - refunds
	total := available + frozen

	return BalanceInfo{
		AvailableCents: available,
		FrozenCents:    frozen,
		TotalCents:     total,
	}, nil
}

// computeBalanceInTx calculates the current balance within a transaction.
// Uses pg_advisory_xact_lock to serialize concurrent balance computations.
func (fp *FundPool) computeBalanceInTx(tx *sql.Tx) (BalanceInfo, error) {
	// Advisory lock scoped to this transaction; released on commit/rollback
	if _, err := tx.Exec(`SELECT pg_advisory_xact_lock(1001)`); err != nil {
		return BalanceInfo{}, fmt.Errorf("failed to acquire advisory lock: %w", err)
	}

	var deposits, freezes, unfreezes, settles, refunds int64
	err := tx.QueryRow(`
		SELECT
			COALESCE(SUM(CASE WHEN operation = 'DEPOSIT' THEN amount_cents ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN operation = 'FREEZE' THEN amount_cents ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN operation = 'UNFREEZE' THEN amount_cents ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN operation = 'SETTLE' THEN amount_cents ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN operation = 'REFUND' THEN amount_cents ELSE 0 END), 0)
		FROM fund_pool_ledger`,
	).Scan(&deposits, &freezes, &unfreezes, &settles, &refunds)
	if err != nil {
		return BalanceInfo{}, fmt.Errorf("failed to compute balance in tx: %w", err)
	}

	frozen := freezes - unfreezes - settles
	available := deposits - freezes + unfreezes - refunds
	total := available + frozen

	return BalanceInfo{
		AvailableCents: available,
		FrozenCents:    frozen,
		TotalCents:     total,
	}, nil
}
