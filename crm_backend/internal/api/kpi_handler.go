package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"crm_backend/internal/repository"
	"github.com/go-chi/chi/v5"
)

// KPIHandler — admin-доступ к KPI менеджеров приёма.
type KPIHandler struct {
	repo *repository.KPIRepository
}

func NewKPIHandler(repo *repository.KPIRepository) *KPIHandler {
	return &KPIHandler{repo: repo}
}

// Routes:
//   GET    /kpi/users/{id}/stats?period=30   — активность пользователя за N дней
//   GET    /kpi/users/{id}/targets           — все цели
//   PUT    /kpi/users/{id}/targets/{metric}  — body { target_count, period_days }
//   DELETE /kpi/targets/{id}                 — удалить цель
func (h *KPIHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/users/{id}/stats", h.stats)
	r.Get("/users/{id}/targets", h.listTargets)
	r.Put("/users/{id}/targets/{metric}", h.upsertTarget)
	r.Delete("/targets/{id}", h.deleteTarget)
	return r
}

func (h *KPIHandler) stats(w http.ResponseWriter, r *http.Request) {
	uid, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, 400, "bad user id")
		return
	}
	period, _ := strconv.Atoi(r.URL.Query().Get("period"))
	if period <= 0 {
		period = 30
	}
	stats, err := h.repo.GetStats(r.Context(), uid, period)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, stats)
}

func (h *KPIHandler) listTargets(w http.ResponseWriter, r *http.Request) {
	uid, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, 400, "bad user id")
		return
	}
	list, err := h.repo.ListTargets(r.Context(), uid)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, list)
}

func (h *KPIHandler) upsertTarget(w http.ResponseWriter, r *http.Request) {
	uid, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, 400, "bad user id")
		return
	}
	metric := chi.URLParam(r, "metric")
	var body struct {
		TargetCount int `json:"target_count"`
		PeriodDays  int `json:"period_days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "invalid payload")
		return
	}
	if body.PeriodDays <= 0 {
		body.PeriodDays = 30
	}
	if err := h.repo.UpsertTarget(r.Context(), uid, metric, body.TargetCount, body.PeriodDays); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (h *KPIHandler) deleteTarget(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, 400, "bad id")
		return
	}
	if err := h.repo.DeleteTarget(r.Context(), id); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}
