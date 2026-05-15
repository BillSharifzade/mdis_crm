-- Wave 2: KPI targets per user + structured call log on interactions

-- Цели KPI для менеджеров приёма (число обработок/созданий за период).
-- period_days = окно (например 30 = за последние 30 дней).
CREATE TABLE IF NOT EXISTS kpi_targets (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    metric VARCHAR(32) NOT NULL,         -- 'processed' | 'created'
    target_count INTEGER NOT NULL DEFAULT 0,
    period_days INTEGER NOT NULL DEFAULT 30,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, metric)
);
CREATE INDEX IF NOT EXISTS idx_kpi_targets_user ON kpi_targets(user_id);

-- Расширение interactions для рефакторинга звонков (T9):
--   outcome   — 'answered' / 'no_answer' / NULL
--   duration_minutes — заполняется, если outcome=answered
ALTER TABLE interactions ADD COLUMN IF NOT EXISTS outcome VARCHAR(20);
ALTER TABLE interactions ADD COLUMN IF NOT EXISTS duration_minutes INTEGER;

-- Индекс на created_by + created_at для быстрого KPI-отчёта
CREATE INDEX IF NOT EXISTS idx_interactions_by_user_time ON interactions(created_by, created_at);
CREATE INDEX IF NOT EXISTS idx_leads_assignee_created ON leads(assignee_id, created_at);

-- Социальная сеть на leads теперь обязательно покрыта (000009),
-- но проверим idempотентно ещё раз.
ALTER TABLE leads ADD COLUMN IF NOT EXISTS social_url VARCHAR(500);
