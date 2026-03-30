-- 002_bidding.up.sql
-- Supplier quotes table

CREATE TABLE supplier_quotes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    trace_id VARCHAR(64) NOT NULL,
    supplier VARCHAR(64) NOT NULL,
    package_title VARCHAR(256) NOT NULL,
    destination VARCHAR(128) NOT NULL,
    duration_days INT NOT NULL,
    duration_nights INT NOT NULL,
    base_price_cents BIGINT NOT NULL,
    refund_guarantee_fee_cents BIGINT NOT NULL DEFAULT 10000,
    total_price_cents BIGINT NOT NULL,
    star_rating DECIMAL(2,1) DEFAULT 0,
    review_count INT DEFAULT 0,
    hotel_name VARCHAR(256),
    highlights JSONB DEFAULT '[]'::jsonb,
    inclusions JSONB DEFAULT '[]'::jsonb,
    image_url VARCHAR(512),
    rank INT,
    is_best_value BOOLEAN DEFAULT FALSE,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_quotes_session ON supplier_quotes(session_id);
CREATE INDEX idx_quotes_session_rank ON supplier_quotes(session_id, rank);
