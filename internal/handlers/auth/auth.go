package auth

import (
	"go-web-starter/internal/handlers"
	"go-web-starter/internal/service"
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
