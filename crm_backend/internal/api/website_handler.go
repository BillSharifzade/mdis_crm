package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"

	"crm_backend/internal/model"
	"crm_backend/internal/repository"
	"crm_backend/internal/service"
	"github.com/go-chi/chi/v5"
)

// WebsiteHandler — публичный (без JWT) приём заявок с лендинга MDIS.
// Лиды помечаются источником "Website" (см. миграцию 000011_website_source.sql).
type WebsiteHandler struct {
	leadSvc  service.ILeadService
	intSvc   service.IInteractionService
	settings *repository.SettingsRepository
}

func NewWebsiteHandler(leadSvc service.ILeadService, intSvc service.IInteractionService, settings *repository.SettingsRepository) *WebsiteHandler {
	return &WebsiteHandler{leadSvc: leadSvc, intSvc: intSvc, settings: settings}
}

func (h *WebsiteHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/lead", h.createLead)
	return r
}

type websiteLeadRequest struct {
	Name    string `json:"name" example:"Алия Каримова"`
	Phone   string `json:"phone" example:"+992900112233"`
	Email   string `json:"email,omitempty" example:"alia@example.com"`
	Message string `json:"message,omitempty" example:"Интересует MBA"`
}

var websitePhoneCleanRx = regexp.MustCompile(`[^\d+]`)
var websiteEmailRx = regexp.MustCompile(`^[\w.+-]+@[\w-]+\.[\w.-]+$`)

// createLead godoc
// @Summary Public website lead intake
// @Description Принимает заявку с сайта MDIS. Публичный, без авторизации (rate-limited по IP). Лид создаётся с источником "Website"; если передано поле message — добавляется в историю как note.
// @Tags integrations
// @Accept json
// @Produce json
// @Param payload body websiteLeadRequest true "Website lead"
// @Success 201 {object} model.Lead
// @Failure 400 {object} map[string]string
// @Router /integrations/website/lead [post]
func (h *WebsiteHandler) createLead(w http.ResponseWriter, r *http.Request) {
	var req websiteLeadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	name := strings.TrimSpace(req.Name)
	rawPhone := strings.TrimSpace(req.Phone)
	email := strings.TrimSpace(req.Email)
	message := strings.TrimSpace(req.Message)

	if name == "" {
		writeJSONError(w, http.StatusBadRequest, "name is required")
		return
	}
	phone := websitePhoneCleanRx.ReplaceAllString(rawPhone, "")
	if len(phone) < 4 {
		writeJSONError(w, http.StatusBadRequest, "valid phone is required")
		return
	}
	if email != "" && !websiteEmailRx.MatchString(email) {
		writeJSONError(w, http.StatusBadRequest, "email is malformed")
		return
	}

	firstName, lastName := splitFullName(name)

	sourceID := h.resolveWebsiteSourceID(r.Context())

	createReq := &model.CreateLeadRequest{
		FirstName: firstName,
		LastName:  lastName,
		Phone:     phone,
		Email:     email,
		SourceID:  sourceID,
		UTMSource: "website",
		UTMMedium: "form",
	}

	lead, err := h.leadSvc.CreateLeadFromForm(r.Context(), createReq)
	if err != nil {
		log.Printf("website lead create error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Failed to create lead")
		return
	}

	if message != "" {
		_, ierr := h.intSvc.AddInteraction(r.Context(), &model.CreateInteractionRequest{
			LeadID:    &lead.ID,
			Type:      "note",
			Direction: "inbound",
			Content:   message,
		})
		if ierr != nil {
			log.Printf("website lead note error: %v", ierr)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(lead)
}

// resolveWebsiteSourceID ищет id источника "Website". Если справочник недоступен
// или такого источника нет — возвращает nil, и сервис подставит дефолтный sourceID.
func (h *WebsiteHandler) resolveWebsiteSourceID(ctx context.Context) *int {
	if h.settings == nil {
		return nil
	}
	sources, err := h.settings.ListSources(ctx, false)
	if err != nil {
		return nil
	}
	for _, s := range sources {
		if strings.EqualFold(s.Name, "Website") {
			id := s.ID
			return &id
		}
	}
	return nil
}

func splitFullName(s string) (string, string) {
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}

func writeJSONError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
