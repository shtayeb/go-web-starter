package auth

import "go-htmx-sqlite/internal/handlers"

type AuthHandler struct {
	handler *handlers.Handlers
}

func NewAuthHandler(h *handlers.Handlers) *AuthHandler {
	return &AuthHandler{
		handler: h,
	}
}
