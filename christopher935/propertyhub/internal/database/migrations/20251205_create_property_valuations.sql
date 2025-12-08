-- Create property_valuations table for storing valuation reports
CREATE TABLE IF NOT EXISTS property_valuations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID REFERENCES properties(id) ON DELETE CASCADE,
    
    -- Valuation results
    estimated_value DECIMAL(12,2) NOT NULL,
    value_low DECIMAL(12,2) NOT NULL,
    value_high DECIMAL(12,2) NOT NULL,
    price_per_sqft DECIMAL(8,2),
    confidence DECIMAL(5,2) NOT NULL,
    
    -- Detailed analysis (stored as JSONB for flexibility)
    comparables JSONB,
    adjustments JSONB,
    market_analysis JSONB,
    valuation_factors JSONB,
    recommendations JSONB,
    
    -- Metadata
    requested_by VARCHAR(255),
    model_version VARCHAR(50) DEFAULT 'v1.0',
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_valuations_property_id ON property_valuations(property_id);
CREATE INDEX IF NOT EXISTS idx_valuations_created_at ON property_valuations(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_valuations_confidence ON property_valuations(confidence);

-- Create updated_at trigger
CREATE OR REPLACE FUNCTION update_valuations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_valuations_updated_at
    BEFORE UPDATE ON property_valuations
    FOR EACH ROW
    EXECUTE FUNCTION update_valuations_updated_at();

-- Add comment
COMMENT ON TABLE property_valuations IS 'Stores property valuation reports with CMA analysis';
