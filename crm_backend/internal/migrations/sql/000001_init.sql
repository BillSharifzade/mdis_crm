-- Users & Auth
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'guest',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS programs (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    faculty VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sources (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS pipeline_stages (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    "order" INTEGER NOT NULL,
    is_final BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS leads (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    email VARCHAR(255),
    phone VARCHAR(50),
    source_id INTEGER REFERENCES sources(id),
    program_id INTEGER REFERENCES programs(id),
    status_id INTEGER REFERENCES pipeline_stages(id),
    assignee_id INTEGER REFERENCES users(id),
    utm_source VARCHAR(255),
    utm_medium VARCHAR(255),
    utm_campaign VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS contacts (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100),
    email VARCHAR(255) UNIQUE,
    phone VARCHAR(50) UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS deals (
    id SERIAL PRIMARY KEY,
    contact_id INTEGER NOT NULL REFERENCES contacts(id),
    stage_id INTEGER NOT NULL REFERENCES pipeline_stages(id),
    assignee_id INTEGER REFERENCES users(id),
    source_id INTEGER REFERENCES sources(id),
    refusal_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE leads ADD COLUMN IF NOT EXISTS contact_id INTEGER REFERENCES contacts(id);

CREATE TABLE IF NOT EXISTS interactions (
    id SERIAL PRIMARY KEY,
    lead_id INTEGER REFERENCES leads(id),
    deal_id INTEGER REFERENCES deals(id),
    type VARCHAR(50) NOT NULL,
    content TEXT NOT NULL,
    duration_seconds INTEGER,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Seed pipeline stages (idempotent)
INSERT INTO pipeline_stages (name, "order", is_final)
SELECT v.name, v.ord, v.is_final FROM (VALUES
    ('Новая заявка', 1, false),
    ('Консультация', 2, false),
    ('Сбор документов', 3, false),
    ('Экзамены/тестирование', 4, false),
    ('Оплата/договор', 5, false),
    ('Зачисление', 6, true),
    ('Отказ', 7, true)
) AS v(name, ord, is_final)
WHERE NOT EXISTS (SELECT 1 FROM pipeline_stages s WHERE s."order" = v.ord);

INSERT INTO sources (name) VALUES
    ('site'), ('telegram'), ('whatsapp'), ('vk'), ('yandex_ads'), ('organic'),
    ('Facebook Ads'), ('Google Ads'), ('Telegram Bot'), ('Direct Traffic')
ON CONFLICT (name) DO NOTHING;

INSERT INTO programs (name, faculty)
SELECT v.name, v.fac FROM (VALUES
    ('Business Management', 'Management'),
    ('Finance', 'Economics'),
    ('Computer Science', 'IT'),
    ('Bachelor of Business Management', 'International BBM program'),
    ('BSc (Hons) Computer Science', 'UK Computer Science degree'),
    ('Foundation Year', 'Pre-university preparation')
) AS v(name, fac)
WHERE NOT EXISTS (SELECT 1 FROM programs p WHERE p.name = v.name);

-- Default admin (password: Admin123!)
INSERT INTO users (name, email, password_hash, role) VALUES
('Admin', 'admin@admin.com', '$2a$10$pAZZANjdgeJ57ucBYQ2vB.ffRSV83ggJPSIqbf6SpO2YDyDGV7DSy', 'admin')
ON CONFLICT (email) DO NOTHING;
