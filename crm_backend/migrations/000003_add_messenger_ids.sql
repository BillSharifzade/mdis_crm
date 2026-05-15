-- Add messenger identifiers to contacts
ALTER TABLE contacts ADD COLUMN telegram_id VARCHAR(100);
ALTER TABLE contacts ADD COLUMN whatsapp_id VARCHAR(100);
ALTER TABLE contacts ADD COLUMN vk_id VARCHAR(100);

-- Create indexes for faster lookup
CREATE INDEX idx_contacts_telegram_id ON contacts(telegram_id);
CREATE INDEX idx_contacts_whatsapp_id ON contacts(whatsapp_id);
CREATE INDEX idx_contacts_vk_id ON contacts(vk_id);
