package api

import (
	"net/http"

	"crm_backend/internal/repository"
)

// StagesPublicHandler — read-only список этапов воронки для всех залогиненных
// пользователей. Полный CRUD у admin под /settings/stages.
type StagesPublicHandler struct {
	repo *repository.SettingsRepository
}

func NewStagesPublicHandler(repo *repository.SettingsRepository) *StagesPublicHandler {
	return &StagesPublicHandler{repo: repo}
}

func (h *StagesPublicHandler) List(w http.ResponseWriter, r *http.Request) {
	onlyActive := r.URL.Query().Get("all") != "1"
	out, err := h.repo.ListStages(r.Context(), onlyActive)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, out)
}
