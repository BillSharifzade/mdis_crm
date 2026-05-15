package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"crm_backend/internal/model"
	"crm_backend/internal/service"
	"crm_backend/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

type UserHandler struct {
	svc service.IUserService
}

func NewUserHandler(svc service.IUserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.listUsers)
	r.Post("/", h.createUser)
	r.Put("/{id}", h.updateUser)
	r.Delete("/{id}", h.deleteUser)
	return r
}

// listUsers godoc
// @Summary List system users
// @Description Get all registered users (admins, admissions, guests).
// @Tags users
// @Produce json
// @Success 200 {array} model.User
// @Security Bearer
// @Router /users [get]
func (h *UserHandler) listUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.ListUsers(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to list users"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// createUser godoc
// @Summary Create a new user
// @Description Register a new system user with a role.
// @Tags users
// @Accept json
// @Produce json
// @Param user body model.CreateUserRequest true "New User"
// @Success 201 {object} model.User
// @Failure 400 {object} map[string]string
// @Security Bearer
// @Router /users [post]
func (h *UserHandler) createUser(w http.ResponseWriter, r *http.Request) {
	var req model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
		return
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Name, email and password are required"})
		return
	}

	validRoles := map[string]bool{"admin": true, "admissions": true, "guest": true}
	if !validRoles[req.Role] {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Role must be admin, admissions, or guest"})
		return
	}

	user, err := h.svc.CreateUser(r.Context(), &req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create user"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// updateUser godoc
// @Summary Update an existing user
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body model.UpdateUserRequest true "User update"
// @Success 200 {object} model.User
// @Security Bearer
// @Router /users/{id} [put]
func (h *UserHandler) updateUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid id"})
		return
	}
	var req model.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid payload"})
		return
	}
	if req.Name == "" || req.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Name and email are required"})
		return
	}
	validRoles := map[string]bool{"admin": true, "admissions": true, "guest": true}
	if !validRoles[req.Role] {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Role must be admin, admissions, or guest"})
		return
	}
	user, err := h.svc.UpdateUser(r.Context(), id, &req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update user"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// deleteUser godoc
// @Summary Delete a user
// @Tags users
// @Param id path int true "User ID"
// @Success 204 "No Content"
// @Security Bearer
// @Router /users/{id} [delete]
func (h *UserHandler) deleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid id"})
		return
	}
	// Админ не может удалить собственный аккаунт.
	if claims, ok := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims); ok {
		if uidFloat, ok := claims["user_id"].(float64); ok && int(uidFloat) == id {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "Нельзя удалить собственный аккаунт"})
			return
		}
	}
	if err := h.svc.DeleteUser(r.Context(), id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete user"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
