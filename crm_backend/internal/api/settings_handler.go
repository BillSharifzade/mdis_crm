package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"crm_backend/internal/model"
	"crm_backend/internal/repository"
	"github.com/go-chi/chi/v5"
)

type SettingsHandler struct {
	repo *repository.SettingsRepository
	bot  *repository.BotSettingsRepository
}

func NewSettingsHandler(repo *repository.SettingsRepository, bot *repository.BotSettingsRepository) *SettingsHandler {
	return &SettingsHandler{repo: repo, bot: bot}
}

// Routes монтируется на /settings. Endpoints:
//
//	GET    /programs       — список факультетов
//	POST   /programs       — создать
//	PUT    /programs/{id}  — изменить
//	DELETE /programs/{id}  — удалить (или soft-archive)
//	GET    /sources        — список источников
//	POST   /sources
//	PUT    /sources/{id}
//	DELETE /sources/{id}
//	GET    /stages         — этапы воронки
//	POST   /stages
//	PUT    /stages/{id}
//	DELETE /stages/{id}
//	POST   /stages/reorder — массовое изменение порядка
//	GET    /bot            — все настройки бота
//	PUT    /bot/{key}      — записать значение
func (h *SettingsHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/programs", h.listPrograms)
	r.Post("/programs", h.createProgram)
	r.Put("/programs/{id}", h.updateProgram)
	r.Delete("/programs/{id}", h.deleteProgram)

	r.Get("/sources", h.listSources)
	r.Post("/sources", h.createSource)
	r.Put("/sources/{id}", h.updateSource)
	r.Delete("/sources/{id}", h.deleteSource)

	r.Get("/stages", h.listStages)
	r.Post("/stages", h.createStage)
	r.Put("/stages/{id}", h.updateStage)
	r.Delete("/stages/{id}", h.deleteStage)
	r.Post("/stages/reorder", h.reorderStages)

	r.Get("/bot", h.listBot)
	r.Put("/bot/{key}", h.setBot)
	return r
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func parseID(r *http.Request) (int, error) {
	return strconv.Atoi(chi.URLParam(r, "id"))
}

// programs ────────────────────────────────────────────────

func (h *SettingsHandler) listPrograms(w http.ResponseWriter, r *http.Request) {
	onlyActive := r.URL.Query().Get("all") != "1"
	out, err := h.repo.ListPrograms(r.Context(), onlyActive)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (h *SettingsHandler) createProgram(w http.ResponseWriter, r *http.Request) {
	var b model.Program
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeErr(w, 400, "invalid payload")
		return
	}
	b.Name = strings.TrimSpace(b.Name)
	if b.Name == "" {
		writeErr(w, 400, "name required")
		return
	}
	p, err := h.repo.CreateProgram(r.Context(), b.Name, b.Faculty)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 201, p)
}

func (h *SettingsHandler) updateProgram(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeErr(w, 400, "bad id")
		return
	}
	var b struct {
		Name     string `json:"name"`
		Faculty  string `json:"faculty"`
		IsActive *bool  `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeErr(w, 400, "invalid payload")
		return
	}
	if err := h.repo.UpdateProgram(r.Context(), id, b.Name, b.Faculty, b.IsActive); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (h *SettingsHandler) deleteProgram(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeErr(w, 400, "bad id")
		return
	}
	if err := h.repo.DeleteProgram(r.Context(), id); err != nil {
		// in-use → archived. сообщаем 200 + archived=true, чтобы фронт показал тост
		writeJSON(w, 200, map[string]interface{}{"status": "archived", "reason": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

// sources ────────────────────────────────────────────────

func (h *SettingsHandler) listSources(w http.ResponseWriter, r *http.Request) {
	onlyActive := r.URL.Query().Get("all") != "1"
	out, err := h.repo.ListSources(r.Context(), onlyActive)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (h *SettingsHandler) createSource(w http.ResponseWriter, r *http.Request) {
	var b model.Source
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeErr(w, 400, "invalid payload")
		return
	}
	b.Name = strings.TrimSpace(b.Name)
	if b.Name == "" {
		writeErr(w, 400, "name required")
		return
	}
	s, err := h.repo.CreateSource(r.Context(), b.Name)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 201, s)
}

func (h *SettingsHandler) updateSource(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeErr(w, 400, "bad id")
		return
	}
	var b struct {
		Name     string `json:"name"`
		IsActive *bool  `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeErr(w, 400, "invalid payload")
		return
	}
	if err := h.repo.UpdateSource(r.Context(), id, b.Name, b.IsActive); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (h *SettingsHandler) deleteSource(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeErr(w, 400, "bad id")
		return
	}
	if err := h.repo.DeleteSource(r.Context(), id); err != nil {
		writeJSON(w, 200, map[string]interface{}{"status": "archived", "reason": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

// stages ────────────────────────────────────────────────

func (h *SettingsHandler) listStages(w http.ResponseWriter, r *http.Request) {
	onlyActive := r.URL.Query().Get("all") != "1"
	out, err := h.repo.ListStages(r.Context(), onlyActive)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (h *SettingsHandler) createStage(w http.ResponseWriter, r *http.Request) {
	var b model.PipelineStage
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeErr(w, 400, "invalid payload")
		return
	}
	b.Name = strings.TrimSpace(b.Name)
	if b.Name == "" {
		writeErr(w, 400, "name required")
		return
	}
	s, err := h.repo.CreateStage(r.Context(), b.Name, b.IsFinal)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 201, s)
}

func (h *SettingsHandler) updateStage(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeErr(w, 400, "bad id")
		return
	}
	var b struct {
		Name     string `json:"name"`
		Order    *int   `json:"order"`
		IsFinal  *bool  `json:"is_final"`
		IsActive *bool  `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeErr(w, 400, "invalid payload")
		return
	}
	if err := h.repo.UpdateStage(r.Context(), id, b.Name, b.Order, b.IsFinal, b.IsActive); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (h *SettingsHandler) reorderStages(w http.ResponseWriter, r *http.Request) {
	var b []model.PipelineStage
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeErr(w, 400, "invalid payload")
		return
	}
	if err := h.repo.ReorderStages(r.Context(), b); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (h *SettingsHandler) deleteStage(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeErr(w, 400, "bad id")
		return
	}
	if err := h.repo.DeleteStage(r.Context(), id); err != nil {
		writeJSON(w, 200, map[string]interface{}{"status": "archived", "reason": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

// bot settings ─────────────────────────────────────────

func (h *SettingsHandler) listBot(w http.ResponseWriter, r *http.Request) {
	m, err := h.bot.GetAll(r.Context())
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, m)
}

func (h *SettingsHandler) setBot(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	var b struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeErr(w, 400, "invalid payload")
		return
	}
	if err := h.bot.Set(r.Context(), key, b.Value); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}
