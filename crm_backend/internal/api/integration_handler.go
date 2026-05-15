package api

import (
	"encoding/json"
	"net/http"

	"crm_backend/internal/model"
	"crm_backend/internal/service"
	"github.com/go-chi/chi/v5"
)

type IntegrationHandler struct {
	svc service.IIntegrationService
}

func NewIntegrationHandler(svc service.IIntegrationService) *IntegrationHandler {
	return &IntegrationHandler{svc: svc}
}

func (h *IntegrationHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/telegram/webhook", h.handleTelegram)
	r.Post("/telephony/webhook", h.handleTelephony)
	return r
}

// handleTelegram godoc
// @Summary Telegram Webhook
// @Description Inbound endpoint for Telegram Bot updates. Maps chats to Leads.
// @Tags integrations
// @Accept json
// @Param payload body model.TelegramWebhookRequest true "Telegram Update"
// @Success 200 "OK"
// @Router /integrations/telegram/webhook [post]
func (h *IntegrationHandler) handleTelegram(w http.ResponseWriter, r *http.Request) {
	var payload model.TelegramWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid payload"})
		return
	}

	err := h.svc.ProcessTelegramWebhook(r.Context(), &payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to handle webhook"})
		return
	}

	w.WriteHeader(http.StatusOK)
}

// handleTelephony godoc
// @Summary Telephony Webhook
// @Description Endpoint for cloud telephony event logs (calls, recordings).
// @Tags integrations
// @Accept json
// @Param payload body model.TelephonyWebhookRequest true "Call Data"
// @Success 200 "OK"
// @Router /integrations/telephony/webhook [post]
func (h *IntegrationHandler) handleTelephony(w http.ResponseWriter, r *http.Request) {
	var payload model.TelephonyWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid payload"})
		return
	}

	err := h.svc.ProcessTelephonyWebhook(r.Context(), &payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to handle telephony webhook"})
		return
	}

	w.WriteHeader(http.StatusOK)
}
