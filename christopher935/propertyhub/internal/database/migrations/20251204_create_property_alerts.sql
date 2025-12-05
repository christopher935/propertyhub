-- Migration: Create property alerts tables
-- Date: 2025-12-04
-- Description: Email alerts for new properties matching user criteria

CREATE TABLE IF NOT EXISTS alert_preferences (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    session_id VARCHAR(255),
    min_price DECIMAL(12,2) DEFAULT 0,
    max_price DECIMAL(12,2) DEFAULT 0,
    min_bedrooms INTEGER DEFAULT 0,
    max_bedrooms INTEGER DEFAULT 0,
    min_bathrooms DECIMAL(3,1) DEFAULT 0,
    preferred_cities TEXT,
    preferred_zips TEXT,
    property_types TEXT,
    alert_frequency VARCHAR(50) DEFAULT 'instant',
    active BOOLEAN DEFAULT true,
    last_notified TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT unique_email_alert UNIQUE (email)
);

CREATE TABLE IF NOT EXISTS property_alerts (
    id BIGSERIAL PRIMARY KEY,
    property_id BIGINT NOT NULL,
    alert_preference_id BIGINT NOT NULL,
    email VARCHAR(255) NOT NULL,
    match_score DECIMAL(5,2) DEFAULT 0,
    sent BOOLEAN DEFAULT false,
    sent_at TIMESTAMP,
    opened BOOLEAN DEFAULT false,
    clicked BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_alert_property FOREIGN KEY (property_id) REFERENCES properties(id) ON DELETE CASCADE,
    CONSTRAINT fk_alert_preference FOREIGN KEY (alert_preference_id) REFERENCES alert_preferences(id) ON DELETE CASCADE
);

CREATE INDEX idx_alert_preferences_email ON alert_preferences(email);
CREATE INDEX idx_alert_preferences_active ON alert_preferences(active) WHERE active = true;
CREATE INDEX idx_property_alerts_property ON property_alerts(property_id);
CREATE INDEX idx_property_alerts_email ON property_alerts(email);
CREATE INDEX idx_property_alerts_sent ON property_alerts(sent, sent_at);

COMMENT ON TABLE alert_preferences IS 'Consumer property alert subscription preferences';
COMMENT ON TABLE property_alerts IS 'Property alert notifications sent to consumers';
