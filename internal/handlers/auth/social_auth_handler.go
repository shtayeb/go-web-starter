package auth

import (
	"context"
	"errors"
	"fmt"
	"go-web-starter/internal/queries"
	"net/http"
	"slices"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

const (
	defaultRedirectURL = "/dashboard"
	// csrfTokenLength    = 32
	// csrfCookieName     = "oauth_csrf"
	maxEmailLength = 255
	maxNameLength  = 100
)

func (ah *AuthHandler) SocialAuthHandler(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")

	// Validate provider
	if !ah.isValidProvider(provider) {
		ah.handler.Logger.PrintInfo("Invalid provider requested", map[string]string{
			"provider": provider,
			"ip":       r.RemoteAddr,
		})
		http.Error(w, "Invalid provider", http.StatusBadRequest)
		return
	}

	// Store intended destination if provided
	if next := r.URL.Query().Get("next"); next != "" && IsValidRedirectPath(next) {
		ah.handler.SessionManager.Put(r.Context(), "next", next)
	}

	gothic.BeginAuthHandler(w, r)
}

func (ah *AuthHandler) SocialAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")

	// Validate provider
	if !ah.isValidProvider(provider) {
		ah.handler.Logger.PrintInfo("Invalid provider in callback", map[string]string{
			"provider": provider,
			"ip":       r.RemoteAddr,
		})
		http.Error(w, "Invalid provider", http.StatusBadRequest)
		return
	}

	// Complete OAuth authentication
	gothUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		ah.handler.Logger.PrintError(err, map[string]string{
			"provider": provider,
		})
		ah.handleAuthError(w, r, "Authentication failed. Please try again.")
		return
	}

	// Validate user data from provider
	if err := ah.validateSocialUserData(gothUser); err != nil {
		ah.handler.Logger.PrintError(err, map[string]string{
			"provider": provider,
			"email":    gothUser.Email,
		})
		ah.handleAuthError(w, r, "Invalid user data received from provider.")
		return
	}

	// Process social authentication
	user, err := ah.authService.ProcessSocialAuth(r.Context(), gothUser, provider)
	if err != nil {
		ah.handler.Logger.PrintError(err, map[string]string{
			"provider": provider,
			"email":    gothUser.Email,
		})
		ah.handleAuthError(w, r, "Login failed. Please try again.")
		return
	}

	// Create new session
	if err := ah.createAuthenticatedSession(r.Context(), user); err != nil {
		ah.handler.ServerError(w, err)
		return
	}

	// Log successful authentication
	ah.handler.Logger.PrintInfo("Successful social login", map[string]string{
		"user_id":  fmt.Sprintf("%d", user.ID),
		"provider": provider,
		"ip":       r.RemoteAddr,
	})

	ah.redirectAfterAuth(w, r)
}

// Helper methods

func (ah *AuthHandler) isValidProvider(provider string) bool {
	// TODO: Make this configurable via environment variables
	validProviders := []string{"google", "github"}
	return slices.Contains(validProviders, provider)
}

func (ah *AuthHandler) validateSocialUserData(user goth.User) error {
	if user.Email == "" {
		return errors.New("email is required")
	}

	if len(user.Email) > maxEmailLength {
		return errors.New("email too long")
	}

	if len(user.Name) > maxNameLength {
		return errors.New("name too long")
	}

	// Validate email format
	if !IsValidEmail(user.Email) {
		return errors.New("invalid email format")
	}

	return nil
}

func (ah *AuthHandler) createAuthenticatedSession(ctx context.Context, user *queries.User) error {
	// Renew session token to prevent fixation
	if err := ah.handler.SessionManager.RenewToken(ctx); err != nil {
		return err
	}

	// Store user ID in session
	ah.handler.SessionManager.Put(ctx, "authenticatedUserID", user.ID)
	ah.handler.SessionManager.Put(ctx, "authenticatedAt", time.Now().Unix())

	return nil
}

func (ah *AuthHandler) redirectAfterAuth(w http.ResponseWriter, r *http.Request) {
	redirectURL := defaultRedirectURL

	// Get intended destination from session (set before OAuth flow)
	if intended := ah.handler.SessionManager.GetString(r.Context(), "next"); intended != "" {
		if IsValidRedirectPath(intended) {
			redirectURL = intended
		}
		ah.handler.SessionManager.Remove(r.Context(), "next")
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (ah *AuthHandler) handleAuthError(w http.ResponseWriter, r *http.Request, userMessage string) {
	ah.handler.SessionManager.Put(r.Context(), "flash", userMessage)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
