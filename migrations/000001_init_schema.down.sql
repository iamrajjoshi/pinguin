-- Drop indexes first
DROP INDEX IF EXISTS idx_checks_monitor_time;

-- Drop tables in reverse order (due to foreign key dependencies)
DROP TABLE IF EXISTS checks;
DROP TABLE IF EXISTS monitors;

-- Drop the extension
DROP EXTENSION IF EXISTS timescaledb;
