# MDIS CRM

CRM-система приёмной комиссии MDIS — учёт лидов, воронка, Telegram-бот для абитуриентов.

## Быстрый старт (одна команда)

```bash
make up
```

Откройте:
- **Frontend** — http://localhost:5173
- **Backend** — http://localhost:8082/api/v1/health
- **Swagger** — http://localhost:8082/api/v1/swagger/

**Логин по умолчанию:** `admin@admin.com` / `Admin123!`

## Состав стека

| Сервис | Порт хоста | Что |
|---|---|---|
| `db` | 5434 | PostgreSQL 15 |
| `backend` | 8082 | Go API (Chi) |
| `frontend` | 5173 | Nginx + Vite build (React 19) |

Миграции применяются автоматически при старте `backend` (idempotent, на основе таблицы `schema_migrations`).

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
├── docker-compose.yml        # ← оркестрация всего
├── Makefile                  # ← удобные команды
├── .env / .env.example       # ← секреты
├── crm_backend/              # Go API (Chi + pgx + JWT)
│   ├── cmd/server/           # точка входа
│   ├── internal/             # api / service / repository / model / migrations
│   ├── pkg/                  # database / middleware / telegram client
│   └── Dockerfile
└── crm-frontend/             # React 19 + Vite 8
    ├── src/                  # pages / components / services
    ├── Dockerfile            # multi-stage build → Nginx serve
    └── nginx.conf
```

## Разработка без Docker

**Backend:**
```bash
cd crm_backend
# поднять только БД
docker compose -f ../docker-compose.yml up -d db
# скопировать env-пример
cp .env.example .env  # или просто:
export DATABASE_URL='postgres://crm_user:crm_password@localhost:5434/mdis_crm?sslmode=disable'
export TELEGRAM_BOT_TOKEN='...'
go run ./cmd/server
```

**Frontend:**
```bash
cd crm-frontend
echo 'VITE_API_BASE_URL=http://localhost:8081/api/v1' > .env  # если backend на 8081
npm install
npm run dev
```

## Документация ТЗ

`crm_backend/tz_mdis_crm.md` — актуальное ТЗ (17.04.2026). Остальные `.md`-файлы устарели.
