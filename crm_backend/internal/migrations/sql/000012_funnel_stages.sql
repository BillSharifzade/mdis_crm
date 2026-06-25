-- 000012: Финализация воронки — переименование и переупорядочивание этапов.
--
-- Итоговый порядок и названия:
--   1  Обращение
--   2  Новая заявка
--   3  Консультация
--   4  EET                        (было «Экзамены/тестирование»)
--   5  Сбор документов
--   6  Оплата/Заключение договора (было «Оплата/договор»)
--   7  Зачисление                 (финальный)
--   8  Отказ                      (финальный)
--
-- Идентификаторы этапов НЕ меняются — существующие leads.status_id и
-- deals.stage_id сохраняют свои ссылки. Меняются только name/"order"/is_final.
-- Сидовые этапы из 000001 имеют id 1..7 (проверено и на проде).

UPDATE pipeline_stages SET name = 'Новая заявка',               "order" = 2, is_final = FALSE, is_active = TRUE, updated_at = NOW() WHERE id = 1;
UPDATE pipeline_stages SET name = 'Консультация',               "order" = 3, is_final = FALSE, is_active = TRUE, updated_at = NOW() WHERE id = 2;
UPDATE pipeline_stages SET name = 'Сбор документов',            "order" = 5, is_final = FALSE, is_active = TRUE, updated_at = NOW() WHERE id = 3;
UPDATE pipeline_stages SET name = 'EET',                        "order" = 4, is_final = FALSE, is_active = TRUE, updated_at = NOW() WHERE id = 4;
UPDATE pipeline_stages SET name = 'Оплата/Заключение договора', "order" = 6, is_final = FALSE, is_active = TRUE, updated_at = NOW() WHERE id = 5;
UPDATE pipeline_stages SET name = 'Зачисление',                 "order" = 7, is_final = TRUE,  is_active = TRUE, updated_at = NOW() WHERE id = 6;
UPDATE pipeline_stages SET name = 'Отказ',                      "order" = 8, is_final = TRUE,  is_active = TRUE, updated_at = NOW() WHERE id = 7;

-- «Обращение» — первый этап. На проде он уже добавлен через Settings UI
-- (в нижнем регистре). Нормализуем существующую запись, иначе вставляем новую.
UPDATE pipeline_stages SET name = 'Обращение', "order" = 1, is_final = FALSE, is_active = TRUE, updated_at = NOW()
 WHERE lower(name) = 'обращение';

INSERT INTO pipeline_stages (name, "order", is_final, is_active)
SELECT 'Обращение', 1, FALSE, TRUE
 WHERE NOT EXISTS (SELECT 1 FROM pipeline_stages WHERE lower(name) = 'обращение');
