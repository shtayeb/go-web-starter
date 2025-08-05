package auth

import (
	"net/http"
)

func (ah *AuthHandler) LogoutPostHandler(w http.ResponseWriter, r *http.Request) {
	err := ah.handler.SessionManager.RenewToken(r.Context())
	if err != nil {
		ah.handler.ServerError(w, err)
		return
	}

	ah.handler.SessionManager.Remove(r.Context(), "authenticatedUserID")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
