-- Ensure the program names used by the Russian-speaking UI/bot exist.
-- Idempotent — won't insert duplicates if the migration is re-run.

INSERT INTO programs (name, faculty)
SELECT v.name, v.fac FROM (VALUES
    ('Бизнес-администрирование', 'Бизнес и менеджмент'),
    ('Информационные технологии', 'IT'),
    ('Финансы и банковское дело', 'Экономика'),
    ('Маркетинг', 'Бизнес и менеджмент'),
    ('Право', 'Юриспруденция'),
    ('Бухгалтерский учёт', 'Экономика')
) AS v(name, fac)
WHERE NOT EXISTS (SELECT 1 FROM programs p WHERE p.name = v.name);
