-- 001_initial.up.sql
-- Core tables: sessions, identity_records, date_configs

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trace_id VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'created',
    -- NLP parsed fields (populated after /nlp/confirm)
    destination VARCHAR(128),
    start_date DATE,
    end_date DATE,
    budget_cents BIGINT,
    adults INT DEFAULT 0,
    children INT DEFAULT 0,
    preferences JSONB DEFAULT '[]'::jsonb,
    is_peak_season BOOLEAN DEFAULT FALSE,
    peak_type VARCHAR(16),
    raw_input TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_trace_id ON sessions(trace_id);
CREATE INDEX idx_sessions_status ON sessions(status);

CREATE TABLE identity_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    trace_id VARCHAR(64) NOT NULL,
    -- All PII fields are AES-256-GCM encrypted
    name_encrypted BYTEA NOT NULL,
    id_number_encrypted BYTEA NOT NULL,
    phone_encrypted BYTEA NOT NULL,
    nonce BYTEA NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_identity_session ON identity_records(session_id);
CREATE INDEX idx_identity_expires ON identity_records(expires_at);

CREATE TABLE date_configs (
    id SERIAL PRIMARY KEY,
    year INT NOT NULL,
    peak_type VARCHAR(16) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    source VARCHAR(32) NOT NULL DEFAULT 'lunar-go',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(year, peak_type)
);
