package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"crm_backend/internal/model"
	"crm_backend/internal/service"

	"github.com/go-chi/chi/v5"
)

type InteractionHandler struct {
	svc service.IInteractionService
}

func NewInteractionHandler(svc service.IInteractionService) *InteractionHandler {
	return &InteractionHandler{svc: svc}
}

func (h *InteractionHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.addInteraction)
	r.Get("/lead/{leadID}", h.getLeadInteractions)
	return r
}

// addInteraction godoc
// @Summary Record a new interaction
// @Description Log a call, SMS, or message against a lead or deal.
// @Tags interactions
// @Accept json
// @Produce json
// @Param interaction body model.CreateInteractionRequest true "Interaction Data"
// @Success 201 {object} model.Interaction
// @Security Bearer
// @Router /interactions [post]
func (h *InteractionHandler) addInteraction(w http.ResponseWriter, r *http.Request) {
	var req model.CreateInteractionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid payload"})
		return
	}

	if req.LeadID == nil && req.DealID == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Either act on a lead or a deal (lead_id or deal_id required)"})
		return
	}

	if req.Type == "" || req.Content == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Type and Content are required fields"})
		return
	}

	interaction, err := h.svc.AddInteraction(r.Context(), &req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to log interaction"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(interaction)
}

// getLeadInteractions godoc
// @Summary Get interaction history for a lead
// @Description Retrieve a list of all calls, messages, and internal notes for a specific lead.
// @Tags interactions
// @Produce json
// @Param leadID path int true "Lead ID"
// @Success 200 {array} model.Interaction
// @Security Bearer
// @Router /interactions/lead/{leadID} [get]
func (h *InteractionHandler) getLeadInteractions(w http.ResponseWriter, r *http.Request) {
	leadIDStr := chi.URLParam(r, "leadID")
	leadID, err := strconv.Atoi(leadIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid lead ID"})
		return
	}

	interactions, err := h.svc.GetLeadHistory(r.Context(), leadID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get interactions"})
		return
	}

	if interactions == nil {
		interactions = []model.Interaction{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(interactions)
}
