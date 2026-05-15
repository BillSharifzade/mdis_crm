-- Audit trail for lead status changes
CREATE TABLE IF NOT EXISTS status_history (
    id SERIAL PRIMARY KEY,
    lead_id INTEGER NOT NULL REFERENCES leads(id) ON DELETE CASCADE,
    old_status_id INTEGER REFERENCES pipeline_stages(id),
    new_status_id INTEGER NOT NULL REFERENCES pipeline_stages(id),
    changed_by INTEGER REFERENCES users(id),
    refusal_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_status_history_lead_id ON status_history(lead_id);
CREATE INDEX IF NOT EXISTS idx_status_history_created_at ON status_history(created_at);
