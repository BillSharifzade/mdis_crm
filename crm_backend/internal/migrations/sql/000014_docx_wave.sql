-- 000014: Доработки CRM по «Технические_комментарии_для_доработки_CRM.docx».
--
--   #1 Статус оплаты (Self Payment / Sponsorship) — отдельное поле лида,
--      НЕ этап воронки. Заодно деактивируем ошибочно заведённые этапы
--      «Self Payment» / «Sponsorship», если админ их создавал через Settings.
--   #2 Новые программы: English for Professionals, Other.
--   #3 Календарное напоминание о повторной коммуникации.
--   #5 Новый источник обращения «Кампус-визит».
--   #6 Доп. поля для программы MBA: место работы (компания / должность).
--   #7 Архив зачисленных — фиксируем момент зачисления (enrolled_at).

-- #1 payment status
ALTER TABLE leads ADD COLUMN IF NOT EXISTS payment_status VARCHAR(20) NOT NULL DEFAULT '';

-- #3 reminder (дата/время повторной коммуникации + флаги доставки/выполнения)
ALTER TABLE leads ADD COLUMN IF NOT EXISTS reminder_at       TIMESTAMPTZ;
ALTER TABLE leads ADD COLUMN IF NOT EXISTS reminder_note     TEXT;
ALTER TABLE leads ADD COLUMN IF NOT EXISTS reminder_done     BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE leads ADD COLUMN IF NOT EXISTS reminder_notified BOOLEAN NOT NULL DEFAULT FALSE;

-- #6 MBA fields
ALTER TABLE leads ADD COLUMN IF NOT EXISTS work_company  VARCHAR(255);
ALTER TABLE leads ADD COLUMN IF NOT EXISTS work_position VARCHAR(255);

-- #7 enrollment archive marker
ALTER TABLE leads ADD COLUMN IF NOT EXISTS enrolled_at TIMESTAMPTZ;

-- Задним числом проставляем enrolled_at уже зачисленным (status_id = 6),
-- чтобы они сразу попали в архив (по дате последнего обновления).
UPDATE leads SET enrolled_at = updated_at WHERE status_id = 6 AND enrolled_at IS NULL;

-- #2 новые программы
INSERT INTO programs (name, faculty)
SELECT v.name, v.fac FROM (VALUES
    ('English for Professionals', 'Foundation'),
    ('Other', '')
) AS v(name, fac)
WHERE NOT EXISTS (SELECT 1 FROM programs p WHERE p.name = v.name);

-- #5 новый источник
INSERT INTO sources (name) VALUES ('Кампус-визит') ON CONFLICT (name) DO NOTHING;

-- #1 убираем способы оплаты из воронки, если они попали в этапы
UPDATE pipeline_stages SET is_active = FALSE, updated_at = NOW()
 WHERE lower(name) IN ('self payment', 'self-payment', 'selfpayment', 'sponsorship');
