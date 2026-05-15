-- Wave 1 schema additions:
--   1. social_url on leads (vk/instagram/facebook link, non-telegram)
--   2. updated_at on programs/sources (used by CRUD for cache busting)
--   3. is_active flag on programs/sources/pipeline_stages — soft-hide instead of delete
--      когда лиды уже ссылаются на запись.
--   4. bot_settings — настраиваемые тексты телеграм-бота (ключ/значение)
--   5. семя новых 7 программ MDIS (англоязычные названия из ТЗ)

ALTER TABLE leads ADD COLUMN IF NOT EXISTS social_url VARCHAR(500);

ALTER TABLE programs ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE programs ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;

ALTER TABLE sources ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE sources ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;

ALTER TABLE pipeline_stages ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE pipeline_stages ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;

CREATE TABLE IF NOT EXISTS bot_settings (
    key VARCHAR(64) PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Default bot copy. Идемпотентно — не перезатираем то, что админ уже поправил.
INSERT INTO bot_settings (key, value) VALUES
    ('greeting', 'Здравствуйте, {name}! 👋

Я бот приёмной комиссии MDIS Tashkent. Чем могу помочь?'),
    ('ask_name', 'Как вас зовут? Напишите, пожалуйста, ФИО.'),
    ('ask_program', 'Спасибо! Какая программа вас интересует? Выберите вариант на кнопке ниже 👇'),
    ('ask_phone', 'Отлично! Оставьте, пожалуйста, ваш номер телефона для связи (например, +992 90 123 45 67).'),
    ('handoff', 'Спасибо! Ваша заявка принята. ✅
Менеджер свяжется с вами в течение 15 минут прямо здесь в чате.'),
    ('faq_admission', 'Условия поступления в MDIS:
• Аттестат о среднем образовании
• Сертификат английского (IELTS 5.5 / TOEFL 46+) или сдача EET
• Заполненная анкета и копия паспорта

Подробнее: https://mdis.uz/admissions'),
    ('faq_event', 'Расскажите, пожалуйста, на какое мероприятие вы хотите зарегистрироваться (Open Day, мастер-класс и т.д.). Менеджер ответит в течение 15 минут.'),
    ('faq_eet', 'EET (English Entry Test) проводится по средам и субботам в 10:00.
Регистрация бесплатная. Менеджер пришлёт расписание и пригласит на ближайшую дату.')
ON CONFLICT (key) DO NOTHING;

-- Новый список программ из ТЗ
INSERT INTO programs (name, faculty)
SELECT v.name, v.fac FROM (VALUES
    ('Professional Certificate in English (PCIE)', 'Foundation'),
    ('BSc (Hons) Cyber Security', 'IT'),
    ('BSc (Hons) Information Technology', 'IT'),
    ('BA (Hons) Business and Financial Management', 'Business'),
    ('BA (Hons) Business and Marketing Management', 'Business'),
    ('BSc (Hons) International Tourism and Hospitality Management (Top-Up)', 'Tourism'),
    ('Masters of Business Administration (MBA)', 'Business')
) AS v(name, fac)
WHERE NOT EXISTS (SELECT 1 FROM programs p WHERE p.name = v.name);

-- Скрываем старые семена, чтобы не путать пользователя.
-- Записи остаются в БД, но не показываются в UI / боте.
UPDATE programs SET is_active = FALSE
WHERE name IN (
    'Business Management', 'Finance', 'Computer Science',
    'Bachelor of Business Management', 'BSc (Hons) Computer Science', 'Foundation Year',
    'Бизнес-администрирование', 'Информационные технологии',
    'Финансы и банковское дело', 'Маркетинг', 'Право', 'Бухгалтерский учёт'
);
