-- 003_locks.up.sql
-- Lock sessions, saga steps, and fund pool ledger tables

-- Lock sessions table for Saga state machine
CREATE TABLE lock_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES sessions(id),
    trace_id VARCHAR(64) NOT NULL,
    quote_id UUID NOT NULL REFERENCES supplier_quotes(id),
    state VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    -- States: PENDING -> LOCKING -> LOCKED -> PAYING -> PAID -> COMPLETED
    -- Failure: LOCK_FAILED, RELEASING, RELEASED, COMPENSATING, COMPENSATED
    locked_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    released_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_lock_sessions_session ON lock_sessions(session_id);
CREATE INDEX idx_lock_sessions_state ON lock_sessions(state);
CREATE INDEX idx_lock_sessions_expires ON lock_sessions(expires_at);

-- Saga steps execution log
CREATE TABLE saga_steps (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    lock_session_id UUID NOT NULL REFERENCES lock_sessions(id),
    trace_id VARCHAR(64) NOT NULL,
    step_name VARCHAR(64) NOT NULL,
    step_order INT NOT NULL,
    direction VARCHAR(16) NOT NULL DEFAULT 'forward',  -- forward or compensate
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    error_message TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_saga_steps_lock ON saga_steps(lock_session_id);

-- Risk control fund pool ledger
CREATE TABLE fund_pool_ledger (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trace_id VARCHAR(64) NOT NULL,
    session_id UUID REFERENCES sessions(id),
    operation VARCHAR(32) NOT NULL,  -- DEPOSIT, FREEZE, UNFREEZE, SETTLE, REFUND
    amount_cents BIGINT NOT NULL,
    balance_after_cents BIGINT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_fund_pool_session ON fund_pool_ledger(session_id);
CREATE INDEX idx_fund_pool_operation ON fund_pool_ledger(operation);
