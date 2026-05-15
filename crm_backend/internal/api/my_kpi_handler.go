package api

import (
	"net/http"
	"strconv"

	"crm_backend/internal/repository"
	"crm_backend/pkg/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

// MyKpiHandler — endpoint, где залогиненный пользователь (admissions/admin)
// смотрит ТОЛЬКО свою активность. Без admin-guard.
type MyKpiHandler struct {
	repo *repository.KPIRepository
}

func NewMyKpiHandler(repo *repository.KPIRepository) *MyKpiHandler {
	return &MyKpiHandler{repo: repo}
}

func (h *MyKpiHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.myStats)
	r.Get("/targets", h.myTargets)
	return r
}

func (h *MyKpiHandler) myTargets(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	if !ok {
		writeErr(w, 401, "no auth")
		return
	}
	uidFloat, _ := claims["user_id"].(float64)
	uid := int(uidFloat)
	if uid <= 0 {
		writeErr(w, 401, "no user")
		return
	}
	list, err := h.repo.ListTargets(r.Context(), uid)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, list)
}

func (h *MyKpiHandler) myStats(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	if !ok {
		writeErr(w, 401, "no auth")
		return
	}
	uidFloat, _ := claims["user_id"].(float64)
	uid := int(uidFloat)
	if uid <= 0 {
		writeErr(w, 401, "no user")
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
