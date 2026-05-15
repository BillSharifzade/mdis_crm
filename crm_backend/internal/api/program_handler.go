package api

import (
	"encoding/json"
	"net/http"

	"crm_backend/internal/model"
	"crm_backend/internal/repository"
	"github.com/go-chi/chi/v5"
)

type ProgramHandler struct {
	repo *repository.ProgramRepository
}

func NewProgramHandler(repo *repository.ProgramRepository) *ProgramHandler {
	return &ProgramHandler{repo: repo}
}

func (h *ProgramHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.list)
	return r
}

// list godoc
// @Summary List academic programs
// @Description Returns all programs (faculties) available for lead intake.
// @Tags programs
// @Produce json
// @Success 200 {array} model.Program
// @Security Bearer
// @Router /programs [get]
func (h *ProgramHandler) list(w http.ResponseWriter, r *http.Request) {
	programs, err := h.repo.ListAll(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to load programs"})
		return
	}
	if programs == nil {
		programs = []model.Program{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(programs)
}
