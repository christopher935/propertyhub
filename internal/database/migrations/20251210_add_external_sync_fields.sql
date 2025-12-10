-- Migration: Add external sync fields to properties table for AppFolio integration
-- Date: 2025-12-10

ALTER TABLE properties ADD COLUMN IF NOT EXISTS external_id VARCHAR(255);
ALTER TABLE properties ADD COLUMN IF NOT EXISTS external_source VARCHAR(50);
ALTER TABLE properties ADD COLUMN IF NOT EXISTS last_synced_at TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_properties_external_id ON properties(external_id);
CREATE INDEX IF NOT EXISTS idx_properties_external_source ON properties(external_source);
CREATE INDEX IF NOT EXISTS idx_properties_external_id_source ON properties(external_id, external_source);
