.PHONY: help up down logs ps rebuild restart clean be-shell db-shell tg-webhook tg-webhook-delete dev-backend dev-backend-stop dev-up

# Подхватываем переменные из .env, если он есть
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

help: ## Показать список команд
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS=":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

up: ## Запустить весь стек (db + backend + frontend) в фоне
	docker compose up -d --build
	@echo ""
	@echo "  ✅ Готово. Открывайте:"
	@echo "     Frontend  → http://localhost:5173"
	@echo "     Backend   → http://localhost:8082/api/v1/health"
	@echo "     Swagger   → http://localhost:8082/api/v1/swagger/"
	@echo ""
	@echo "  Логин по умолчанию: admin@admin.com / Admin123!"

down: ## Остановить стек
	docker compose down

logs: ## Показать логи (Ctrl+C для выхода)
	docker compose logs -f --tail=100

ps: ## Статус сервисов
	docker compose ps

rebuild: ## Пересобрать образы без кэша
	docker compose build --no-cache
	docker compose up -d

restart: ## Перезапустить backend и frontend (без БД)
	docker compose restart backend frontend

clean: ## Снести всё, включая volume БД (⚠️ удалит данные)
	docker compose down -v

be-shell: ## Зайти в shell внутри backend контейнера
	docker compose exec backend sh

db-shell: ## Открыть psql внутри БД
	docker compose exec db psql -U crm_user -d mdis_crm

tg-webhook: ## Зарегистрировать Telegram webhook (нужен PUBLIC_URL через ngrok)
	@test -n "$(PUBLIC_URL)" || (echo "Укажите PUBLIC_URL=https://your-ngrok.app"; exit 1)
	@test -n "$(TELEGRAM_BOT_TOKEN)" || (echo "TELEGRAM_BOT_TOKEN не задан (положите его в .env)"; exit 1)
	@echo "→ setWebhook на $(PUBLIC_URL)/api/v1/integrations/telegram/webhook"
	@curl -sS -F "url=$(PUBLIC_URL)/api/v1/integrations/telegram/webhook" \
	  "https://api.telegram.org/bot$(TELEGRAM_BOT_TOKEN)/setWebhook" && echo ""
	@echo "→ getWebhookInfo:"
	@curl -sS "https://api.telegram.org/bot$(TELEGRAM_BOT_TOKEN)/getWebhookInfo" && echo ""

tg-webhook-delete: ## Удалить Telegram webhook
	@test -n "$(TELEGRAM_BOT_TOKEN)" || (echo "TELEGRAM_BOT_TOKEN не задан"; exit 1)
	@curl -sS "https://api.telegram.org/bot$(TELEGRAM_BOT_TOKEN)/deleteWebhook" && echo ""

# ── Аварийный режим, если docker daemon чудит (см. ошибку TTRPC/Yunix) ─────
# Поднимаем только БД и фронт в докере, а Go-бэкенд запускаем на хосте.
# Используется ./crm_backend на хосте, порт 8082 такой же, как в docker-compose.

dev-up: ## Поднять DB+frontend в docker, бэкенд оставить для dev-backend
	docker compose up -d db frontend

dev-backend: ## Собрать и запустить бэкенд на хосте (порт 8082, БД из docker)
	@pkill -f "/tmp/mdis_server" 2>/dev/null || true
	@cd crm_backend && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /tmp/mdis_server ./cmd/server
	@echo "→ запускаю /tmp/mdis_server на :8082 (лог: /tmp/mdis_server.log)"
	@PORT=8082 \
	  DATABASE_URL='postgres://crm_user:crm_password@localhost:5434/mdis_crm?sslmode=disable' \
	  TELEGRAM_BOT_TOKEN='$(TELEGRAM_BOT_TOKEN)' \
	  JWT_SECRET='$(JWT_SECRET)' \
	  WEBHOOK_SECRET='$(WEBHOOK_SECRET)' \
	  nohup /tmp/mdis_server > /tmp/mdis_server.log 2>&1 &
	@sleep 2 && curl -sS http://localhost:8082/api/v1/health && echo ""
	@echo "✅ Бэкенд запущен. Остановить:  make dev-backend-stop"

dev-backend-stop: ## Остановить host-бэкенд
	@pkill -f "/tmp/mdis_server" && echo "stopped" || echo "не запущен"
