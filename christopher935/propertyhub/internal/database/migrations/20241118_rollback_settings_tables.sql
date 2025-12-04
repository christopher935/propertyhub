-- Rollback: Drop user_profiles and user_preferences tables
-- Date: 2024-11-18
-- Description: Rolls back the settings tables migration

BEGIN;

-- Drop triggers first
DROP TRIGGER IF EXISTS update_user_profiles_updated_at ON user_profiles;
DROP TRIGGER IF EXISTS update_user_preferences_updated_at ON user_preferences;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables (CASCADE will remove foreign key constraints)
DROP TABLE IF EXISTS user_preferences CASCADE;
DROP TABLE IF EXISTS user_profiles CASCADE;

COMMIT;

-- Verify rollback
SELECT 'Rollback completed successfully' AS status;
