-- Telegram bot dialog state and message direction
-- ТЗ от 13.04.2026: «Лиды из Telegram: первоначально в контакте с ботом,
-- далее в том же чате подключается менеджер»

ALTER TABLE interactions
    ADD COLUMN IF NOT EXISTS direction VARCHAR(16) NOT NULL DEFAULT 'inbound';

-- Привязка Telegram-чата к лиду + состояние сценария бота
CREATE TABLE IF NOT EXISTS telegram_chats (
    id SERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL UNIQUE,
    lead_id INTEGER REFERENCES leads(id) ON DELETE CASCADE,
    bot_state VARCHAR(32) NOT NULL DEFAULT 'greet',   -- greet → ask_name → ask_program → ask_phone → manager
    bot_active BOOLEAN NOT NULL DEFAULT TRUE,         -- false = менеджер перехватил диалог
    tg_username VARCHAR(128),
    tg_first_name VARCHAR(128),
    collected_name TEXT,
    collected_program TEXT,
    collected_phone TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_telegram_chats_lead_id ON telegram_chats(lead_id);
CREATE INDEX IF NOT EXISTS idx_telegram_chats_chat_id ON telegram_chats(chat_id);
