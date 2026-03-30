-- 005_refunds.up.sql
-- Refund request tracking with state machine

CREATE TABLE refund_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id),
    session_id UUID NOT NULL REFERENCES sessions(id),
    trace_id VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    -- States: pending -> processing -> supplier_refunded -> pool_compensated -> user_refunded / failed
    total_refund_cents BIGINT NOT NULL,
    supplier_recoverable_cents BIGINT DEFAULT 0,
    pool_compensation_cents BIGINT DEFAULT 0,
    supplier_refund_at TIMESTAMPTZ,
    user_refund_at TIMESTAMPTZ,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refund_order ON refund_requests(order_id);
CREATE INDEX idx_refund_session ON refund_requests(session_id);
