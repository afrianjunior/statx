-- Up migration: Create account table

CREATE TABLE account (
    id TEXT PRIMARY KEY,
    username VARCHAR NOT NULL UNIQUE,
    password VARCHAR NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_login DATETIME
);

CREATE INDEX idx_account_username ON account(username);
CREATE INDEX idx_account_last_login ON account(last_login);

-- Create a trigger to ensure unique IDs
CREATE TRIGGER tr_account_generate_uuid
AFTER INSERT ON account
FOR EACH ROW
WHEN NEW.id IS NULL
BEGIN
   UPDATE account SET id = (
     lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || 
     substr(lower(hex(randomblob(2))),2) || '-' || 
     substr('89ab',abs(random()) % 4 + 1, 1) || 
     substr(lower(hex(randomblob(2))),2) || '-' || 
     lower(hex(randomblob(6)))
   ) WHERE rowid = NEW.rowid;
END;

