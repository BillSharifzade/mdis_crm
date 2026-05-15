-- Create app_state table to store round-robin counter and other system state
CREATE TABLE IF NOT EXISTS app_state (
    key VARCHAR(50) PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Initialize round-robin counter
INSERT INTO app_state (key, value) VALUES ('last_assigned_user_id', '0') ON CONFLICT DO NOTHING;
