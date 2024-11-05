-- Down migration: Drop config_monitor table
DROP TRIGGER IF EXISTS tr_config_monitor_generate_uuid;
DROP TABLE IF EXISTS config_monitor;
DROP INDEX IF EXISTS idx_config_monitor_type;
DROP INDEX IF EXISTS idx_config_monitor_name;
