-- Migration: Create sync_statuses table for AppFolio sync tracking
-- Date: 2025-12-10

CREATE TABLE IF NOT EXISTS sync_statuses (
    id SERIAL PRIMARY KEY,
    sync_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    properties_synced INTEGER DEFAULT 0,
    properties_created INTEGER DEFAULT 0,
    properties_updated INTEGER DEFAULT 0,
    properties_deleted INTEGER DEFAULT 0,
    error_count INTEGER DEFAULT 0,
    errors TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sync_statuses_sync_type ON sync_statuses(sync_type);
CREATE INDEX IF NOT EXISTS idx_sync_statuses_created_at ON sync_statuses(created_at);
