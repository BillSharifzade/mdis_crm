package api

import (
	"net/http"

	"crm_backend/internal/repository"
)

// SourcesPublicHandler — read-only список источников обращения для всех
// залогиненных пользователей (нужен формам создания/редактирования лида).
// Полный CRUD у admin под /settings/sources.
type SourcesPublicHandler struct {
	repo *repository.SettingsRepository
}

func NewSourcesPublicHandler(repo *repository.SettingsRepository) *SourcesPublicHandler {
	return &SourcesPublicHandler{repo: repo}
}

func (h *SourcesPublicHandler) List(w http.ResponseWriter, r *http.Request) {
	onlyActive := r.URL.Query().Get("all") != "1"
	out, err := h.repo.ListSources(r.Context(), onlyActive)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, out)
}
