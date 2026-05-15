-- Support for Tasks within Interactions
ALTER TABLE interactions ADD COLUMN is_task BOOLEAN DEFAULT FALSE;
ALTER TABLE interactions ADD COLUMN is_done BOOLEAN DEFAULT FALSE;
ALTER TABLE interactions ADD COLUMN due_date TIMESTAMP WITH TIME ZONE;
