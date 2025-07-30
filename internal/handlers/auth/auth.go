package auth

import (
	"go-htmx-sqlite/internal/handlers"
	"go-htmx-sqlite/internal/service"
)

type AuthHandler struct {
	handler     *handlers.Handlers
	authService *service.AuthService
}

func NewAuthHandler(h *handlers.Handlers, authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		handler:     h,
		authService: authService,
	}
}
