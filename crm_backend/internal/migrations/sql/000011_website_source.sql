-- Источник для заявок, пришедших с публичного API сайта MDIS.
INSERT INTO sources (name) VALUES ('Website')
ON CONFLICT (name) DO NOTHING;
