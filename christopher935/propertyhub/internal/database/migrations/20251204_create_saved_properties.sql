-- Migration: Create saved_properties table
-- Date: 2025-12-04
-- Description: Allows consumers to save/favorite properties for later viewing

CREATE TABLE IF NOT EXISTS saved_properties (
    id BIGSERIAL PRIMARY KEY,
    session_id VARCHAR(255) NOT NULL,
    property_id BIGINT NOT NULL,
    email VARCHAR(255),
    saved_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_saved_property FOREIGN KEY (property_id) REFERENCES properties(id) ON DELETE CASCADE,
    CONSTRAINT unique_session_property UNIQUE (session_id, property_id)
);

CREATE INDEX idx_saved_properties_session ON saved_properties(session_id);
CREATE INDEX idx_saved_properties_property ON saved_properties(property_id);
CREATE INDEX idx_saved_properties_email ON saved_properties(email);
CREATE INDEX idx_saved_properties_saved_at ON saved_properties(saved_at DESC);

COMMENT ON TABLE saved_properties IS 'Consumer saved/favorited properties';
COMMENT ON COLUMN saved_properties.session_id IS 'Session identifier for anonymous users';
COMMENT ON COLUMN saved_properties.email IS 'Optional email for logged-in users or email capture';
