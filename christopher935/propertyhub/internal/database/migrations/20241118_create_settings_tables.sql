-- Migration: Create user_profiles and user_preferences tables for admin settings
-- Date: 2024-11-18
-- Description: Adds comprehensive profile and preferences support for admin users

BEGIN;

-- Create user_profiles table
CREATE TABLE IF NOT EXISTS user_profiles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(50),
    company VARCHAR(200),
    department VARCHAR(100),
    job_title VARCHAR(100),
    avatar_url TEXT,
    bio TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);

CREATE INDEX idx_user_profiles_user_id ON user_profiles(user_id);

COMMENT ON TABLE user_profiles IS 'Extended profile information for admin users';
COMMENT ON COLUMN user_profiles.avatar_url IS 'URL to profile photo stored in DigitalOcean Spaces';

-- Create user_preferences table
CREATE TABLE IF NOT EXISTS user_preferences (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    timezone VARCHAR(50) DEFAULT 'America/Chicago',
    language VARCHAR(10) DEFAULT 'en',
    date_format VARCHAR(20) DEFAULT 'MM/DD/YYYY',
    time_format VARCHAR(10) DEFAULT '12h',
    email_notifications BOOLEAN DEFAULT true,
    sms_notifications BOOLEAN DEFAULT false,
    desktop_notifications BOOLEAN DEFAULT true,
    weekly_reports BOOLEAN DEFAULT true,
    theme VARCHAR(20) DEFAULT 'light',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);

CREATE INDEX idx_user_preferences_user_id ON user_preferences(user_id);

COMMENT ON TABLE user_preferences IS 'User preferences and notification settings for admin users';

-- Create default preferences for existing admin users
INSERT INTO user_preferences (user_id, timezone, language, date_format, time_format, email_notifications, sms_notifications, desktop_notifications, weekly_reports, theme)
SELECT 
    id,
    'America/Chicago',
    'en',
    'MM/DD/YYYY',
    '12h',
    true,
    false,
    true,
    true,
    'light'
FROM admin_users
WHERE id NOT IN (SELECT user_id FROM user_preferences);

-- Create trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_user_profiles_updated_at
    BEFORE UPDATE ON user_profiles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_preferences_updated_at
    BEFORE UPDATE ON user_preferences
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMIT;

-- Verify migration
SELECT 'Migration completed successfully' AS status;
SELECT COUNT(*) AS admin_users_count FROM admin_users;
SELECT COUNT(*) AS user_profiles_count FROM user_profiles;
SELECT COUNT(*) AS user_preferences_count FROM user_preferences;
