-- Users & Auth
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'guest', -- admin, admission, guest
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Programs/Faculties (Metadata)
CREATE TABLE programs (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    faculty VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Traffic Sources (Metadata)
CREATE TABLE sources (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL, -- 'site', 'telegram', 'whatsapp', 'vk', 'organic'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Pipeline Stages
CREATE TABLE pipeline_stages (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    "order" INTEGER NOT NULL,
    is_final BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Leads: Initial form submissions, inbound messages
CREATE TABLE leads (
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

-- Contacts: Dedicated potential student records
CREATE TABLE contacts (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100),
    email VARCHAR(255) UNIQUE,
    phone VARCHAR(50) UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Deals/Applications
CREATE TABLE deals (
    id SERIAL PRIMARY KEY,
    contact_id INTEGER NOT NULL REFERENCES contacts(id),
    stage_id INTEGER NOT NULL REFERENCES pipeline_stages(id),
    assignee_id INTEGER REFERENCES users(id),
    source_id INTEGER REFERENCES sources(id),
    refusal_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Link Leads to Contacts (A lead can become a contact, or be associated)
ALTER TABLE leads ADD COLUMN contact_id INTEGER REFERENCES contacts(id);

-- Interactions: Calls, Messages, Emails
CREATE TABLE interactions (
    id SERIAL PRIMARY KEY,
    lead_id INTEGER REFERENCES leads(id),
    deal_id INTEGER REFERENCES deals(id),
    type VARCHAR(50) NOT NULL, -- 'call', 'sms', 'email', 'messenger', 'manual_note'
    content TEXT NOT NULL,
    duration_seconds INTEGER, -- for calls
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Initial Data
INSERT INTO pipeline_stages (name, "order", is_final) VALUES 
('Новая заявка', 1, false),
('Консультация', 2, false),
('Сбор документов', 3, false),
('Экзамены/тестирование', 4, false),
('Оплата/договор', 5, false),
('Зачисление', 6, true),
('Отказ', 7, true);

INSERT INTO sources (name) VALUES 
('site'), ('telegram'), ('whatsapp'), ('vk'), ('yandex_ads'), ('organic');

INSERT INTO programs (name, faculty) VALUES 
('Business Management', 'Management'),
('Finance', 'Economics'),
('Computer Science', 'IT');

INSERT INTO programs (name, faculty) VALUES 
('Bachelor of Business Management', 'International BBM program'),
('BSc (Hons) Computer Science', 'UK Computer Science degree'),
('Foundation Year', 'Pre-university preparation');

INSERT INTO sources (name) VALUES 
('Facebook Ads'),
('Google Ads'),
('Telegram Bot'),
('Direct Traffic');

-- Test Admin (password: Admin123!)
INSERT INTO users (name, email, password_hash, role) VALUES 
('Admin', 'admin@admin.com', '$2a$10$pAZZANjdgeJ57ucBYQ2vB.ffRSV83ggJPSIqbf6SpO2YDyDGV7DSy', 'admin');

COMMIT;
