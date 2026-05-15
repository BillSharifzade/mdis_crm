package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"crm_backend/internal/model"
	"crm_backend/internal/repository"
	"crm_backend/internal/service"
	"crm_backend/pkg/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

type TelegramChatHandler struct {
	bot     *service.TelegramBotService
	intSvc  service.IInteractionService
	chatRepo *repository.TelegramChatRepository
}

func NewTelegramChatHandler(bot *service.TelegramBotService, intSvc service.IInteractionService, chatRepo *repository.TelegramChatRepository) *TelegramChatHandler {
	return &TelegramChatHandler{bot: bot, intSvc: intSvc, chatRepo: chatRepo}
}

func (h *TelegramChatHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/unread", h.unread)
	r.Get("/{leadID}/status", h.getStatus)
	r.Get("/{leadID}/messages", h.getMessages)
	r.Post("/{leadID}/send", h.sendMessage)
	r.Post("/{leadID}/takeover", h.takeover)
	return r
}

func (h *TelegramChatHandler) unread(w http.ResponseWriter, r *http.Request) {
	m, err := h.chatRepo.UnreadByLead(r.Context())
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	// Возвращаем массив {lead_id, unread} — JSON-объекты с int ключами не дружат с JS.
	type item struct {
		LeadID int `json:"lead_id"`
		Unread int `json:"unread"`
	}
	out := make([]item, 0, len(m))
	for k, v := range m {
		out = append(out, item{LeadID: k, Unread: v})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func (h *TelegramChatHandler) extractUserID(r *http.Request) int {
	claims, ok := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	if !ok {
		return 0
	}
	if v, ok := claims["user_id"].(float64); ok {
		return int(v)
	}
	return 0
}

func (h *TelegramChatHandler) getStatus(w http.ResponseWriter, r *http.Request) {
	leadID, err := strconv.Atoi(chi.URLParam(r, "leadID"))
	if err != nil {
		http.Error(w, `{"error":"bad leadID"}`, http.StatusBadRequest)
		return
	}
	chat, err := h.bot.GetChatStatus(r.Context(), leadID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	if chat == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"chat": nil, "has_chat": false})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"chat":     chat,
		"has_chat": true,
	})
}

func (h *TelegramChatHandler) getMessages(w http.ResponseWriter, r *http.Request) {
	leadID, err := strconv.Atoi(chi.URLParam(r, "leadID"))
	if err != nil {
		http.Error(w, `{"error":"bad leadID"}`, http.StatusBadRequest)
		return
	}
	all, err := h.intSvc.GetLeadHistory(r.Context(), leadID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	// Возвращаем только messenger-сообщения, отсортированные по возрастанию времени
	out := make([]model.Interaction, 0, len(all))
	for _, it := range all {
		if it.Type == "messenger" {
			out = append(out, it)
		}
	}
	// разворот по времени по возрастанию (ListByLead возвращает DESC)
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func (h *TelegramChatHandler) sendMessage(w http.ResponseWriter, r *http.Request) {
	leadID, err := strconv.Atoi(chi.URLParam(r, "leadID"))
	if err != nil {
		http.Error(w, `{"error":"bad leadID"}`, http.StatusBadRequest)
		return
	}
	var req model.SendTelegramRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"bad payload"}`, http.StatusBadRequest)
		return
	}
	uid := h.extractUserID(r)
	if err := h.bot.ManagerSend(r.Context(), leadID, uid, req.Text); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (h *TelegramChatHandler) takeover(w http.ResponseWriter, r *http.Request) {
	leadID, err := strconv.Atoi(chi.URLParam(r, "leadID"))
	if err != nil {
		http.Error(w, `{"error":"bad leadID"}`, http.StatusBadRequest)
		return
	}
	if err := h.bot.Takeover(r.Context(), leadID); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"ok": true})
}
