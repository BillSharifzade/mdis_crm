package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"crm_backend/internal/service"
	"github.com/go-chi/chi/v5"
)

type StatusHistoryHandler struct {
	svc service.IStatusHistoryService
}

func NewStatusHistoryHandler(svc service.IStatusHistoryService) *StatusHistoryHandler {
	return &StatusHistoryHandler{svc: svc}
}

func (h *StatusHistoryHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/lead/{leadID}", h.getByLead)
	return r
}

// getByLead godoc
// @Summary Get status change history for a lead
// @Description Returns the full audit trail of pipeline stage changes for a specific lead.
// @Tags status-history
// @Produce json
// @Param leadID path int true "Lead ID"
// @Success 200 {array} model.StatusHistory
// @Security Bearer
// @Router /status-history/lead/{leadID} [get]
func (h *StatusHistoryHandler) getByLead(w http.ResponseWriter, r *http.Request) {
	leadIDStr := chi.URLParam(r, "leadID")
	leadID, err := strconv.Atoi(leadIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid lead ID"})
		return
	}

	history, err := h.svc.GetLeadStatusHistory(r.Context(), leadID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to load status history"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
