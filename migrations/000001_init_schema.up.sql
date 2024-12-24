-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Create monitors table
CREATE TABLE monitors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    url TEXT NOT NULL,
    name TEXT NOT NULL,
    interval INTEGER NOT NULL DEFAULT 60, -- in seconds
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    -- TODO: add configuration in the future
);

-- Create checks table (for storing monitoring results)
CREATE TABLE checks (
    time TIMESTAMPTZ NOT NULL,
    monitor_id UUID NOT NULL REFERENCES monitors(id),
    duration_ms INTEGER,
    status_code INTEGER,
    error TEXT,
    success BOOLEAN NOT NULL,
    headers JSONB,
    body TEXT,
    body_size INTEGER,
    PRIMARY KEY (monitor_id, time)
);

-- Convert checks to hypertable
SELECT create_hypertable('checks', 'time');

-- Add indexes
CREATE INDEX idx_checks_monitor_time ON checks (monitor_id, time DESC);
