-- Up migration: Create config_monitor table

CREATE TABLE config_monitor (
    id TEXT PRIMARY KEY,
    type TEXT CHECK(type IN ('uptime', 'generic')),
    method VARCHAR,
    name VARCHAR,
    url VARCHAR,
    interval INTEGER,
    icon VARCHAR,
    color VARCHAR,
    max_retry INTEGER,
    retry_interval INTEGER,
    call_method VARCHAR,
    call_encoding VARCHAR,
    call_body VARCHAR,
    call_headers VARCHAR
);

CREATE INDEX idx_config_monitor_type ON config_monitor(type);
CREATE INDEX idx_config_monitor_name ON config_monitor(name);

-- Create a trigger to ensure unique IDs
CREATE TRIGGER tr_config_monitor_generate_uuid
AFTER INSERT ON config_monitor
FOR EACH ROW
WHEN NEW.id IS NULL
BEGIN
   UPDATE config_monitor SET id = (
     lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || 
     substr(lower(hex(randomblob(2))),2) || '-' || 
     substr('89ab',abs(random()) % 4 + 1, 1) || 
     substr(lower(hex(randomblob(2))),2) || '-' || 
     lower(hex(randomblob(6)))
   ) WHERE rowid = NEW.rowid;
END;

