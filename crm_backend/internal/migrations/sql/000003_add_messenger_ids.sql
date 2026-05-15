ALTER TABLE contacts ADD COLUMN IF NOT EXISTS telegram_id VARCHAR(100);
ALTER TABLE contacts ADD COLUMN IF NOT EXISTS whatsapp_id VARCHAR(100);
ALTER TABLE contacts ADD COLUMN IF NOT EXISTS vk_id VARCHAR(100);

CREATE INDEX IF NOT EXISTS idx_contacts_telegram_id ON contacts(telegram_id);
CREATE INDEX IF NOT EXISTS idx_contacts_whatsapp_id ON contacts(whatsapp_id);
CREATE INDEX IF NOT EXISTS idx_contacts_vk_id ON contacts(vk_id);
