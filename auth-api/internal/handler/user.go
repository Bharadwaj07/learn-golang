package handler

import (
	"auth-api/internal/domain"
	"encoding/json"
	"net/http"
)

// contextKey is a private type for context keys — avoids collisions
type contextKey string

const UserIDKey contextKey = "user_id"

type UserHandler struct {
	svc domain.UserService
}

func NewUserHandler(svc domain.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// Register handles POST /register
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, domain.NewBadRequest("invalid json"))
		return
	}

	user, err := h.svc.Register(r.Context(), body.Email, body.Password)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

// Login handles POST /login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, domain.NewBadRequest("invalid json"))
		return
	}

	token, err := h.svc.Login(r.Context(), body.Email, body.Password)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

// GetProfile handles GET /profile — protected route
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// user_id was injected by AuthMiddleware into context
	userID, ok := r.Context().Value(UserIDKey).(uint)
	if !ok {
		writeError(w, domain.NewUnauthorized("unauthorized"))
		return
	}

	user, err := h.svc.GetProfile(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}
