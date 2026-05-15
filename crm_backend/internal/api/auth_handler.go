package api

import (
	"encoding/json"
	"net/http"

	"crm_backend/internal/model"
	"crm_backend/internal/service"
	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	svc service.IAuthService
}

func NewAuthHandler(svc service.IAuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/login", h.login)
	return r
}

// login godoc
// @Summary Login and get a JWT token
// @Description Authenticate a user with email and password to receive a Bearer token.
// @Tags auth
// @Accept json
// @Produce json
// @Param login body model.LoginRequest true "Login Request"
// @Success 200 {object} model.LoginResponse
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid payload"})
		return
	}

	res, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid credentials"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
