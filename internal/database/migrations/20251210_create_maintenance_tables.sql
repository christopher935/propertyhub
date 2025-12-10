-- Migration: Create maintenance request tables for AppFolio integration
-- Created: 2025-12-10

-- Vendors table
CREATE TABLE IF NOT EXISTS vendors (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    company_name VARCHAR(255),
    category VARCHAR(100) NOT NULL,
    phone VARCHAR(50),
    email VARCHAR(255),
    address TEXT,
    hourly_rate DECIMAL(10, 2) DEFAULT 0,
    minimum_charge DECIMAL(10, 2) DEFAULT 0,
    is_preferred BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    rating DECIMAL(3, 2) DEFAULT 0,
    total_jobs INTEGER DEFAULT 0,
    avg_response_time INTEGER DEFAULT 0,
    notes TEXT,
    license_number VARCHAR(100),
    insurance_expiry TIMESTAMP,
    service_areas JSONB DEFAULT '[]',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_vendors_category ON vendors(category);
CREATE INDEX IF NOT EXISTS idx_vendors_is_preferred ON vendors(is_preferred);
CREATE INDEX IF NOT EXISTS idx_vendors_is_active ON vendors(is_active);

-- Maintenance requests table
CREATE TABLE IF NOT EXISTS maintenance_requests (
    id SERIAL PRIMARY KEY,
    appfolio_id VARCHAR(100) UNIQUE NOT NULL,
    property_id INTEGER,
    tenant_id INTEGER,
    tenant_name VARCHAR(255),
    tenant_phone VARCHAR(50),
    tenant_email VARCHAR(255),
    property_address VARCHAR(500) NOT NULL,
    unit_number VARCHAR(50),
    description TEXT NOT NULL,
    category VARCHAR(100) DEFAULT 'general',
    priority VARCHAR(50) DEFAULT 'medium',
    status VARCHAR(50) DEFAULT 'open',
    suggested_vendor VARCHAR(255),
    assigned_vendor VARCHAR(255),
    assigned_vendor_id INTEGER REFERENCES vendors(id),
    ai_triage_result JSONB,
    estimated_cost DECIMAL(10, 2),
    actual_cost DECIMAL(10, 2),
    response_time VARCHAR(50),
    scheduled_date TIMESTAMP,
    completed_date TIMESTAMP,
    notes TEXT,
    internal_notes TEXT,
    permission_to_enter BOOLEAN DEFAULT FALSE,
    pet_on_premises BOOLEAN DEFAULT FALSE,
    last_synced_at TIMESTAMP,
    resolved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_maintenance_requests_appfolio_id ON maintenance_requests(appfolio_id);
CREATE INDEX IF NOT EXISTS idx_maintenance_requests_status ON maintenance_requests(status);
CREATE INDEX IF NOT EXISTS idx_maintenance_requests_priority ON maintenance_requests(priority);
CREATE INDEX IF NOT EXISTS idx_maintenance_requests_property_id ON maintenance_requests(property_id);
CREATE INDEX IF NOT EXISTS idx_maintenance_requests_tenant_id ON maintenance_requests(tenant_id);
CREATE INDEX IF NOT EXISTS idx_maintenance_requests_assigned_vendor_id ON maintenance_requests(assigned_vendor_id);

-- Maintenance status logs for audit trail
CREATE TABLE IF NOT EXISTS maintenance_status_logs (
    id SERIAL PRIMARY KEY,
    maintenance_request_id INTEGER NOT NULL REFERENCES maintenance_requests(id),
    old_status VARCHAR(50),
    new_status VARCHAR(50) NOT NULL,
    changed_by VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_maintenance_status_logs_request_id ON maintenance_status_logs(maintenance_request_id);

-- Maintenance alerts for emergency notifications
CREATE TABLE IF NOT EXISTS maintenance_alerts (
    id SERIAL PRIMARY KEY,
    maintenance_request_id INTEGER NOT NULL REFERENCES maintenance_requests(id),
    alert_type VARCHAR(50) NOT NULL,
    message TEXT,
    sent_at TIMESTAMP,
    acknowledged_at TIMESTAMP,
    acknowledged_by VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_maintenance_alerts_request_id ON maintenance_alerts(maintenance_request_id);

-- Insert some default vendors for testing
INSERT INTO vendors (name, company_name, category, phone, email, hourly_rate, is_preferred, is_active, rating) VALUES
('Joe''s Plumbing', 'Joe''s Plumbing Services', 'plumbing', '555-0201', 'joe@joesplumbing.com', 85.00, true, true, 4.8),
('Sparky Electric', 'Sparky Electrical Services', 'electrical', '555-0202', 'info@sparkyelectric.com', 95.00, true, true, 4.7),
('Cool Air HVAC', 'Cool Air Heating & Cooling', 'hvac', '555-0203', 'service@coolairhvac.com', 110.00, true, true, 4.9),
('Fix-It-All Appliances', 'Fix-It-All Appliance Repair', 'appliance', '555-0204', 'repair@fixitall.com', 75.00, false, true, 4.5),
('Handy Home Services', 'Handy Home General Services', 'general', '555-0205', 'info@handyhome.com', 65.00, true, true, 4.6),
('Pest Away', 'Pest Away Extermination', 'pest', '555-0206', 'help@pestaway.com', 125.00, true, true, 4.8),
('Structure Pro', 'Structure Pro Construction', 'structural', '555-0207', 'projects@structurepro.com', 150.00, false, true, 4.4)
ON CONFLICT DO NOTHING;
