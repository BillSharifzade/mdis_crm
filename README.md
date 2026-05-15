# MDIS CRM

> CRM-система приёмной комиссии **MDIS** — учёт лидов, воронка продаж, Telegram-бот для абитуриентов и публичный API для заявок с сайта.

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![React](https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=white)](https://react.dev/)
[![Postgres](https://img.shields.io/badge/PostgreSQL-15-336791?logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)

CRM for MDIS admissions: lead pipeline, manager KPI, Telegram-bot intake, public website form intake, Excel import, role-based access (admin / admissions / guest).

---

## Быстрый старт

```bash
git clone https://github.com/BillSharifzade/mdis_crm.git
cd mdis_crm
cp .env.example .env   # отредактируйте секреты
make up
```

Открыть:
- **Frontend** — http://localhost:5173
- **Backend API** — http://localhost:8082/api/v1/health
- **Swagger** — http://localhost:8082/api/v1/swagger/

**Логин по умолчанию:** `admin@admin.com` / `Admin123!`

## Состав стека

| Сервис | Порт хоста | Что |
|---|---|---|
| `db` | 5434 | PostgreSQL 15 |
| `backend` | 8082 | Go API (Chi + pgx + JWT) |
| `frontend` | 5173 | Nginx + Vite build (React 19) |

Миграции применяются автоматически при старте `backend` (idempotent, на основе таблицы `schema_migrations`).

## Возможности

- **Воронка из 7 стадий** (новая → консультация → документы → экзамены → оплата → зачисление; финал — отказ с причиной).
- **Round-robin** распределение лидов между менеджерами приёма.
- **Telegram-бот**: анкета через state-machine (`greet → ask_name → ask_program → ask_phone → manager`), затем handoff на менеджера в том же чате.
- **Публичный API для сайта**: `POST /integrations/website/lead` — приём заявок с лендинга, без авторизации, под rate-limit.
- **Импорт лидов из Excel** (`*.xlsx`, многолистовая книга, авто-маппинг колонок RU/EN).
- **Аналитика**: dashboard, конверсия по стадиям, KPI менеджеров, экспорт XLSX/PDF.
- **SSE-стрим** для real-time обновлений UI (`/api/v1/events`).
- **Роли**: `admin` / `admissions` / `guest`.

## Команды Make

```bash
make help       # показать список
make up         # поднять всё (build + start)
make down       # остановить
make logs       # хвост логов всех сервисов
make ps         # статус
make restart    # рестартнуть только backend+frontend
make rebuild    # пересобрать без кэша
make clean      # снести всё, включая volume БД ⚠️
make be-shell   # shell в backend контейнере
make db-shell   # psql в БД
make tg-webhook PUBLIC_URL=https://your-ngrok.app   # зарегистрировать Telegram webhook
```

## Конфигурация

Скопируйте `.env.example` → `.env` и заполните секреты:

```env
TELEGRAM_BOT_TOKEN=<bot token from @BotFather>
JWT_SECRET=<random long string>
WEBHOOK_SECRET=<for telephony webhook>
# опционально:
SMTP_HOST=...
SMTP_PORT=...
```

`.env` подхватывается `docker compose` автоматически. Без переменных стек тоже поднимется — будут разумные дефолты.

## Публичный API для сайта

Заявка с лендинга MDIS, без авторизации (rate-limit 10/мин с IP):

```http
POST /api/v1/integrations/website/lead
Content-Type: application/json

{
  "name": "Алия Каримова",
  "phone": "+992900112233",
  "email": "alia@example.com",
  "message": "Интересует MBA"
}
```

Ответ `201 Created` с объектом лида. Поля:
- `name` — обязательное
- `phone` — обязательное (≥ 4 цифр после нормализации)
- `email` — опционально (валидируется по формату)
- `message` — опционально, попадёт в историю лида как `note`

Лид помечается источником **`Website`**, `utm_source=website`, `utm_medium=form`, и автоматически назначается на менеджера по round-robin.

## Telegram-бот

1. Создайте бота через `@BotFather` → положите токен в `.env`
2. Откройте внешний туннель: `ngrok http 8082`
3. Зарегистрируйте webhook: `make tg-webhook PUBLIC_URL=https://abc.ngrok-free.app`
4. В Telegram → `/start` → пройти анкету (ФИО → программа → телефон)
5. В CRM → карточка нового лида → кнопка **«Чат в Telegram»**

Сценарий: бот собирает анкету, затем менеджер подключается из CRM в том же чате (handoff).

## Структура проекта

```
mdis_crm/
├── docker-compose.yml        # оркестрация всего
├── Makefile                  # удобные команды
├── .env.example              # шаблон секретов
├── crm_backend/              # Go API (Chi + pgx + JWT)
│   ├── cmd/server/           # точка входа
│   ├── internal/             # api / service / repository / model / migrations
│   ├── pkg/                  # database / middleware / telegram client
│   └── Dockerfile
└── crm-frontend/             # React 19 + Vite
    ├── src/                  # pages / components / services
    ├── Dockerfile            # multi-stage build → Nginx serve
    └── nginx.conf
```

## Разработка без Docker

**Backend:**
```bash
cd crm_backend
docker compose -f ../docker-compose.yml up -d db   # поднять только БД
export DATABASE_URL='postgres://crm_user:crm_password@localhost:5434/mdis_crm?sslmode=disable'
export TELEGRAM_BOT_TOKEN='...'
go run ./cmd/server
```

**Frontend:**
```bash
cd crm-frontend
echo 'VITE_API_BASE_URL=http://localhost:8082/api/v1' > .env
npm install
npm run dev
```

## API Reference

Полная OpenAPI-спецификация доступна по адресу `http://localhost:8082/api/v1/swagger/` после `make up`.

Шорткаты:
| Метод | Путь | Доступ |
|---|---|---|
| `POST` | `/auth/login` | публичный, rate-limit |
| `GET`/`POST` | `/leads` | JWT |
| `POST` | `/leads/import` | JWT, multipart `*.xlsx` |
| `GET` | `/analytics/dashboard` | JWT (admin/guest) |
| `POST` | `/integrations/website/lead` | **публичный**, rate-limit |
| `POST` | `/integrations/telegram/webhook` | публичный, X-Telegram-Bot-Api-Secret-Token |
| `POST` | `/integrations/telephony/webhook` | X-Webhook-Secret |
| `GET` | `/events?token=<jwt>` | SSE-стрим |

## Документация ТЗ

`crm_backend/tz_mdis_crm.md` — актуальное ТЗ. Остальные `.md`-файлы устарели.

## License

Распространяется по лицензии **GNU GPL v3** — см. файл [LICENSE](LICENSE).
