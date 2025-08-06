package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/markbates/goth/gothic"
)

func (ah *AuthHandler) SocialAuthHandler(w http.ResponseWriter, r *http.Request) {
	// try to get the user without re-authenticating
	if gothUser, err := gothic.CompleteUserAuth(w, r); err == nil {
		internalUser, err := ah.authService.SocialLogin(r.Context(), gothUser.Name, gothUser.Email, gothUser.Provider)
		if err != nil {
			ah.handler.ServerError(w, err)
			return
		}

		// Session manager - renew token AFTER successful authentication to prevent session fixation
		err = ah.handler.SessionManager.RenewToken(r.Context())
		if err != nil {
			ah.handler.ServerError(w, err)
			return
		}

		ah.handler.SessionManager.Put(r.Context(), "authenticatedUserID", internalUser.ID)

		redirectURL := "/dashboard"
		nextPath := r.URL.Query().Get("next")

		if nextPath != "" && IsValidRedirectPath(nextPath) {
			redirectURL = nextPath
		}

		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}

	gothic.BeginAuthHandler(w, r)
}

func (ah *AuthHandler) SocialAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// This is after the social login
	provider := chi.URLParam(r, "provider")

	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		ah.handler.ServerError(w, err)
		return
	}

	internalUser, err := ah.authService.SocialSignUp(r.Context(), provider, user.Name, user.Email)
	if err != nil {
		ah.handler.ServerError(w, err)
		return
	}

	// Session manager - renew token AFTER successful authentication to prevent session fixation
	err = ah.handler.SessionManager.RenewToken(r.Context())
	if err != nil {
		ah.handler.ServerError(w, err)
		return
	}

	ah.handler.SessionManager.Put(r.Context(), "authenticatedUserID", internalUser.ID)

	redirectURL := "/dashboard"
	nextPath := r.URL.Query().Get("next")

	if nextPath != "" && IsValidRedirectPath(nextPath) {
		redirectURL = nextPath
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}
