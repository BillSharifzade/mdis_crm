-- 000013: Уровень английского языка абитуриента.
--
-- Одно текстовое поле, хранит выбранную систему теста и балл, например:
--   «EET 3.5», «IELTS 6.0», «DUOLINGO 100». Пусто = не указан.
-- Шкалы (для UI/бота):
--   EET:      1.0 … 10.0 с шагом 0.5
--   IELTS:    5.5, 6.0
--   DUOLINGO: 80, 90, 100, 120

ALTER TABLE leads ADD COLUMN IF NOT EXISTS english_level VARCHAR(50);

-- Бот собирает уровень в анкете до создания лида — храним во временной
-- колонке чата, как collected_name/program/phone.
ALTER TABLE telegram_chats ADD COLUMN IF NOT EXISTS collected_english VARCHAR(50);
