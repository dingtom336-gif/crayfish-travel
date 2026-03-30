-- 004_payments_orders.up.sql
-- Payment records and order management

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES sessions(id),
    lock_session_id UUID REFERENCES lock_sessions(id),
    quote_id UUID NOT NULL REFERENCES supplier_quotes(id),
    trace_id VARCHAR(64) NOT NULL,
    out_trade_no VARCHAR(64) NOT NULL UNIQUE,
    method VARCHAR(32) NOT NULL DEFAULT 'qr',  -- qr, voice_token
    amount_cents BIGINT NOT NULL,
    qr_code_url VARCHAR(512),
    voice_token VARCHAR(128),
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    -- States: pending -> paid -> refunding -> refunded / failed
    alipay_trade_no VARCHAR(64),
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_session ON payments(session_id);
CREATE INDEX idx_payments_out_trade ON payments(out_trade_no);
CREATE INDEX idx_payments_status ON payments(status);

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES sessions(id),
    payment_id UUID NOT NULL REFERENCES payments(id),
    quote_id UUID NOT NULL REFERENCES supplier_quotes(id),
    trace_id VARCHAR(64) NOT NULL,
    order_no VARCHAR(32) NOT NULL UNIQUE,
    status VARCHAR(32) NOT NULL DEFAULT 'created',
    -- States: created -> confirmed -> fulfilling -> completed -> refund_requested -> refunded
    contact_name VARCHAR(128),
    contact_phone VARCHAR(32),
    total_amount_cents BIGINT NOT NULL,
    base_price_cents BIGINT NOT NULL,
    refund_guarantee_fee_cents BIGINT NOT NULL,
    supplier VARCHAR(64) NOT NULL,
    package_title VARCHAR(256) NOT NULL,
    destination VARCHAR(128) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    adults INT NOT NULL DEFAULT 0,
    children INT NOT NULL DEFAULT 0,
    sms_sent BOOLEAN DEFAULT FALSE,
    sms_sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_session ON orders(session_id);
CREATE INDEX idx_orders_order_no ON orders(order_no);
CREATE INDEX idx_orders_status ON orders(status);
