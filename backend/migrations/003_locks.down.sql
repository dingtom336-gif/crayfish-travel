-- 003_locks.down.sql
-- Drop lock sessions, saga steps, and fund pool ledger tables

DROP INDEX IF EXISTS idx_fund_pool_operation;
DROP INDEX IF EXISTS idx_fund_pool_session;
DROP TABLE IF EXISTS fund_pool_ledger;

DROP INDEX IF EXISTS idx_saga_steps_lock;
DROP TABLE IF EXISTS saga_steps;

DROP INDEX IF EXISTS idx_lock_sessions_expires;
DROP INDEX IF EXISTS idx_lock_sessions_state;
DROP INDEX IF EXISTS idx_lock_sessions_session;
DROP TABLE IF EXISTS lock_sessions;
