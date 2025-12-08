-- ============================================================================
-- PropertyHub Database Schema Audit Script
-- Purpose: Audit existing database schema and compare with expected models
-- ============================================================================

-- ============================================================================
-- SECTION 1: LIST ALL EXISTING TABLES
-- ============================================================================
\echo '============================================================================'
\echo 'SECTION 1: ALL EXISTING TABLES IN DATABASE'
\echo '============================================================================'

SELECT 
    schemaname,
    tablename,
    tableowner
FROM pg_tables 
WHERE schemaname = 'public'
ORDER BY tablename;

-- ============================================================================
-- SECTION 2: TABLE DETAILS - COLUMNS, TYPES, AND CONSTRAINTS
-- ============================================================================
\echo ''
\echo '============================================================================'
\echo 'SECTION 2: TABLE COLUMNS, TYPES, AND CONSTRAINTS'
\echo '============================================================================'

SELECT 
    t.table_name,
    c.column_name,
    c.data_type,
    c.character_maximum_length,
    c.column_default,
    c.is_nullable,
    CASE 
        WHEN pk.column_name IS NOT NULL THEN 'PRIMARY KEY'
        WHEN uk.column_name IS NOT NULL THEN 'UNIQUE'
        ELSE ''
    END as constraint_type
FROM information_schema.tables t
JOIN information_schema.columns c 
    ON t.table_name = c.table_name AND t.table_schema = c.table_schema
LEFT JOIN (
    SELECT ku.table_name, ku.column_name
    FROM information_schema.table_constraints tc
    JOIN information_schema.key_column_usage ku 
        ON tc.constraint_name = ku.constraint_name
    WHERE tc.constraint_type = 'PRIMARY KEY'
) pk ON c.table_name = pk.table_name AND c.column_name = pk.column_name
LEFT JOIN (
    SELECT ku.table_name, ku.column_name
    FROM information_schema.table_constraints tc
    JOIN information_schema.key_column_usage ku 
        ON tc.constraint_name = ku.constraint_name
    WHERE tc.constraint_type = 'UNIQUE'
) uk ON c.table_name = uk.table_name AND c.column_name = uk.column_name
WHERE t.table_schema = 'public' 
  AND t.table_type = 'BASE TABLE'
ORDER BY t.table_name, c.ordinal_position;

-- ============================================================================
-- SECTION 3: FOREIGN KEY RELATIONSHIPS
-- ============================================================================
\echo ''
\echo '============================================================================'
\echo 'SECTION 3: FOREIGN KEY RELATIONSHIPS'
\echo '============================================================================'

SELECT 
    tc.table_name AS source_table,
    kcu.column_name AS source_column,
    ccu.table_name AS target_table,
    ccu.column_name AS target_column,
    tc.constraint_name
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu
    ON tc.constraint_name = kcu.constraint_name
JOIN information_schema.constraint_column_usage ccu
    ON ccu.constraint_name = tc.constraint_name
WHERE tc.constraint_type = 'FOREIGN KEY'
ORDER BY tc.table_name;

-- ============================================================================
-- SECTION 4: INDEXES
-- ============================================================================
\echo ''
\echo '============================================================================'
\echo 'SECTION 4: DATABASE INDEXES'
\echo '============================================================================'

SELECT 
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE schemaname = 'public'
ORDER BY tablename, indexname;

-- ============================================================================
-- SECTION 5: TABLE ROW COUNTS
-- ============================================================================
\echo ''
\echo '============================================================================'
\echo 'SECTION 5: TABLE ROW COUNTS'
\echo '============================================================================'

SELECT 
    relname AS table_name,
    n_live_tup AS row_count
FROM pg_stat_user_tables
WHERE schemaname = 'public'
ORDER BY n_live_tup DESC;

-- ============================================================================
-- SECTION 6: EXPECTED VS ACTUAL SCHEMA COMPARISON
-- ============================================================================
\echo ''
\echo '============================================================================'
\echo 'SECTION 6: EXPECTED VS ACTUAL SCHEMA COMPARISON'
\echo '============================================================================'

\echo 'Expected Core Tables (from models):'
\echo '  - properties (Property)'
\echo '  - admin_users (AdminUser)'
\echo '  - bookings (Booking)'
\echo '  - leads (Lead)'
\echo '  - contacts (Contact)'
\echo '  - closing_pipelines (ClosingPipeline)'
\echo '  - webhook_events (WebhookEvent)'
\echo ''
\echo 'Expected Behavioral Tables:'
\echo '  - behavioral_events (BehavioralEvent)'
\echo '  - behavioral_sessions (BehavioralSession)'
\echo '  - behavioral_scores (BehavioralScore)'
\echo ''
\echo 'Expected Application Workflow Tables:'
\echo '  - property_application_groups (PropertyApplicationGroup)'
\echo '  - application_numbers (ApplicationNumber)'
\echo '  - application_applicants (ApplicationApplicant)'
\echo ''
\echo 'Expected Email/Notification Tables:'
\echo '  - email_events (EmailEvent)'
\echo '  - campaigns (Campaign)'
\echo '  - email_batches (EmailBatch)'
\echo '  - email_templates (EmailTemplate)'
\echo '  - trusted_email_senders (TrustedEmailSender)'
\echo '  - email_processing_rules (EmailProcessingRule)'
\echo ''
\echo 'Expected Admin/Security Tables:'
\echo '  - admin_sessions'
\echo '  - security_events'
\echo '  - notification_states (NotificationState)'
\echo '  - admin_notifications (AdminNotification)'
\echo '  - data_imports (DataImport)'
\echo '  - price_change_events (PriceChangeEvent)'

-- Check for missing expected tables
\echo ''
\echo 'MISSING TABLES CHECK:'
SELECT expected_table, 
       CASE WHEN EXISTS (
           SELECT 1 FROM information_schema.tables 
           WHERE table_schema = 'public' AND table_name = expected_table
       ) THEN 'EXISTS' ELSE 'MISSING' END as status
FROM (VALUES 
    ('properties'),
    ('admin_users'),
    ('bookings'),
    ('leads'),
    ('contacts'),
    ('closing_pipelines'),
    ('webhook_events'),
    ('behavioral_events'),
    ('behavioral_sessions'),
    ('behavioral_scores'),
    ('property_application_groups'),
    ('application_numbers'),
    ('application_applicants'),
    ('email_events'),
    ('campaigns'),
    ('email_batches'),
    ('email_templates'),
    ('trusted_email_senders'),
    ('email_processing_rules'),
    ('admin_sessions'),
    ('security_events'),
    ('notification_states'),
    ('admin_notifications'),
    ('data_imports'),
    ('price_change_events')
) AS expected(expected_table)
ORDER BY status DESC, expected_table;

-- ============================================================================
-- SECTION 7: TABLES WITH NO MODEL DEFINITION (POTENTIALLY ORPHANED)
-- ============================================================================
\echo ''
\echo '============================================================================'
\echo 'SECTION 7: TABLES WITH NO KNOWN MODEL (REVIEW NEEDED)'
\echo '============================================================================'

SELECT tablename 
FROM pg_tables 
WHERE schemaname = 'public'
  AND tablename NOT IN (
    'properties',
    'admin_users',
    'bookings',
    'leads',
    'contacts',
    'closing_pipelines',
    'webhook_events',
    'behavioral_events',
    'behavioral_sessions',
    'behavioral_scores',
    'property_application_groups',
    'application_numbers',
    'application_applicants',
    'email_events',
    'campaigns',
    'email_batches',
    'email_templates',
    'trusted_email_senders',
    'email_processing_rules',
    'admin_sessions',
    'security_events',
    'notification_states',
    'admin_notifications',
    'data_imports',
    'price_change_events',
    'schema_migrations'
  )
ORDER BY tablename;

-- ============================================================================
-- SECTION 8: COLUMN TYPE MISMATCHES (Common Issues)
-- ============================================================================
\echo ''
\echo '============================================================================'
\echo 'SECTION 8: POTENTIAL COLUMN TYPE ISSUES'
\echo '============================================================================'

-- Check for TEXT columns that might need to be JSONB
SELECT table_name, column_name, data_type
FROM information_schema.columns
WHERE table_schema = 'public'
  AND (
    (column_name LIKE '%_data' AND data_type = 'text')
    OR (column_name LIKE '%_json' AND data_type = 'text')
    OR (column_name LIKE '%payload%' AND data_type = 'text')
  )
ORDER BY table_name, column_name;

-- Check for VARCHAR columns without length constraints
SELECT table_name, column_name, data_type, character_maximum_length
FROM information_schema.columns
WHERE table_schema = 'public'
  AND data_type = 'character varying'
  AND character_maximum_length IS NULL
ORDER BY table_name, column_name;

\echo ''
\echo '============================================================================'
\echo 'DATABASE AUDIT COMPLETE'
\echo '============================================================================'
