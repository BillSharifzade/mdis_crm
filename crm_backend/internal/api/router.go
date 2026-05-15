package api

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "crm_backend/docs"
	"crm_backend/internal/model"
	"crm_backend/internal/repository"
	"crm_backend/internal/service"
	"crm_backend/pkg/database"
	"crm_backend/pkg/events"
	"crm_backend/pkg/middleware"
)

type Router struct {
	db               *database.DB
	telegramBotToken string
}

func NewRouter(db *database.DB, telegramBotToken string) *Router {
	return &Router{db: db, telegramBotToken: telegramBotToken}
}

func (r *Router) InitRoutes() http.Handler {
	mux := chi.NewRouter()

	// Global middleware (порядок важен: безопасность → логирование → recovery).
	mux.Use(middleware.SecurityHeaders)
	mux.Use(middleware.CORSMiddleware)
	mux.Use(chimiddleware.RequestID)
	mux.Use(chimiddleware.RealIP)
	mux.Use(chimiddleware.Logger)
	mux.Use(chimiddleware.Recoverer)
	// Лимит на тело запроса — защита от DoS большими payload.
	// Multipart (импорт) пропускается через свой ParseMultipartForm.
	mux.Use(middleware.BodySizeLimit(5 << 20)) // 5 MB

	// Rate-limit для /auth/login — 5 попыток за минуту с одного IP.
	loginLimiter := middleware.NewLoginRateLimiter(5, time.Minute)
	// Rate-limit для публичной заявки с сайта — 10 в минуту с IP, защита от спама.
	websiteLeadLimiter := middleware.NewLoginRateLimiter(10, time.Minute)

	// Repositories
	userRepo := repository.NewUserRepository(r.db)
	leadRepo := repository.NewLeadRepository(r.db)
	contactRepo := repository.NewContactRepository(r.db)
	dealRepo := repository.NewDealRepository(r.db)
	interactionRepo := repository.NewInteractionRepository(r.db)
	analyticsRepo := repository.NewAnalyticsRepository(r.db)
	statusHistoryRepo := repository.NewStatusHistoryRepository(r.db)
	tgChatRepo := repository.NewTelegramChatRepository(r.db)
	programRepo := repository.NewProgramRepository(r.db)
	settingsRepo := repository.NewSettingsRepository(r.db)
	botSettingsRepo := repository.NewBotSettingsRepository(r.db)
	kpiRepo := repository.NewKPIRepository(r.db)

	// Real-time events bus (SSE)
	eventBus := events.New()

	// Services
	notificationSvc := service.NewNotificationService()
	authSvc := service.NewAuthService(userRepo)
	leadSvc := service.NewLeadService(leadRepo, contactRepo, dealRepo, userRepo, notificationSvc, statusHistoryRepo, eventBus)
	interactionSvc := service.NewInteractionService(interactionRepo, eventBus)
	analyticsSvc := service.NewAnalyticsService(analyticsRepo)
	botSvc := service.NewTelegramBotService(r.telegramBotToken, tgChatRepo, leadSvc, interactionSvc, contactRepo, leadRepo, programRepo, settingsRepo, botSettingsRepo, eventBus)
	integrationSvc := service.NewIntegrationService(leadSvc, interactionSvc, contactRepo, leadRepo, botSvc)
	exportSvc := service.NewExportService(leadRepo)
	userSvc := service.NewUserService(userRepo)
	statusHistorySvc := service.NewStatusHistoryService(statusHistoryRepo)

	// Cron scheduler
	cronSvc := service.NewCronService(r.db, interactionSvc, notificationSvc, analyticsSvc, statusHistoryRepo)
	cronSvc.Start(context.Background())

	// Handlers
	authHandler := NewAuthHandler(authSvc)
	leadHandler := NewLeadHandler(leadSvc)
	interactionHandler := NewInteractionHandler(interactionSvc)
	analyticsHandler := NewAnalyticsHandler(analyticsSvc, exportSvc)
	integrationHandler := NewIntegrationHandler(integrationSvc)
	userHandler := NewUserHandler(userSvc)
	statusHistoryHandler := NewStatusHistoryHandler(statusHistorySvc)
	tgChatHandler := NewTelegramChatHandler(botSvc, interactionSvc, tgChatRepo)
	programHandler := NewProgramHandler(programRepo)
	settingsHandler := NewSettingsHandler(settingsRepo, botSettingsRepo)
	stagesPublicHandler := NewStagesPublicHandler(settingsRepo)
	importHandler := NewImportHandler(leadSvc, settingsRepo)
	eventsHandler := NewEventsHandler(eventBus)
	kpiHandler := NewKPIHandler(kpiRepo)
	myKpiHandler := NewMyKpiHandler(kpiRepo)
	websiteHandler := NewWebsiteHandler(leadSvc, interactionSvc, settingsRepo)

	mux.Route("/api/v1", func(api chi.Router) {
		// Public: Health check
		api.Get("/health", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		})

		// Public: Swagger docs
		api.Get("/swagger/*", httpSwagger.Handler())

		// Public: Auth (login) — под rate-limiter
		api.With(loginLimiter.Middleware).Mount("/auth", authHandler.Routes())

		// Public SSE stream — EventSource не умеет слать заголовки, поэтому
		// auth идёт через ?token=<jwt> (валидируется внутри хендлера).
		api.Get("/events", eventsHandler.Route)

		// Public Telegram webhook — вызывается серверами Telegram, не из CRM.
		// Если задан TELEGRAM_WEBHOOK_SECRET — проверяем header X-Telegram-Bot-Api-Secret-Token.
		api.Post("/integrations/telegram/webhook", func(w http.ResponseWriter, req *http.Request) {
			expected := os.Getenv("TELEGRAM_WEBHOOK_SECRET")
			if expected != "" {
				got := req.Header.Get("X-Telegram-Bot-Api-Secret-Token")
				if subtle.ConstantTimeCompare([]byte(got), []byte(expected)) != 1 {
					w.WriteHeader(http.StatusForbidden)
					json.NewEncoder(w).Encode(map[string]string{"error": "Forbidden"})
					return
				}
			}
			var payload model.TelegramWebhookRequest
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Invalid payload"})
				return
			}
			if err := integrationSvc.ProcessTelegramWebhook(req.Context(), &payload); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			w.WriteHeader(http.StatusOK)
		})

		// Telephony webhook остаётся за X-Webhook-Secret
		api.Group(func(webhooks chi.Router) {
			webhooks.Use(middleware.WebhookSecretMiddleware)
			webhooks.Mount("/integrations", integrationHandler.Routes())
		})

		// Публичный приём заявок с сайта MDIS — без JWT, под rate-limiter.
		api.With(websiteLeadLimiter.Middleware).Post("/integrations/website/lead", websiteHandler.createLead)

		// Protected routes — require JWT authentication
		api.Group(func(protected chi.Router) {
			protected.Use(middleware.AuthMiddleware)

			protected.Mount("/leads", leadHandler.Routes())
			protected.Mount("/interactions", interactionHandler.Routes())
			protected.Mount("/status-history", statusHistoryHandler.Routes())
			protected.Mount("/programs", programHandler.Routes())
			// Публичный (для всех залогиненных) read-only список этапов воронки.
			// Полный CRUD остаётся под /settings/stages (admin).
			protected.Get("/stages", stagesPublicHandler.List)
			// Чат менеджера с лидом через Telegram (внутри CRM)
			protected.Mount("/telegram-chat", tgChatHandler.Routes())
			// Импорт лидов из Excel
			protected.Mount("/leads-import", importHandler.Routes())

			// Org-аналитика — admin + guest. У admissions есть только свой KPI.
			protected.Group(func(g chi.Router) {
				g.Use(middleware.RoleMiddleware("admin", "guest"))
				g.Mount("/analytics", analyticsHandler.Routes())
			})

			// Свой KPI — admin + admissions.
			protected.Group(func(g chi.Router) {
				g.Use(middleware.RoleMiddleware("admin", "admissions"))
				g.Mount("/me/kpi", myKpiHandler.Routes())
			})

			// Admin-only routes
			protected.Group(func(admin chi.Router) {
				admin.Use(middleware.RoleMiddleware("admin"))
				admin.Mount("/users", userHandler.Routes())
				admin.Mount("/settings", settingsHandler.Routes())
				admin.Mount("/kpi", kpiHandler.Routes())
			})
		})
	})

	return mux
}
