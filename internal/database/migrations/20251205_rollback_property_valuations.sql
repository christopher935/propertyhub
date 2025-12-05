-- Rollback script for property_valuations table
DROP TRIGGER IF EXISTS trigger_valuations_updated_at ON property_valuations;
DROP FUNCTION IF EXISTS update_valuations_updated_at();
DROP TABLE IF EXISTS property_valuations CASCADE;
