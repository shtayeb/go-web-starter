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

	ah.handler.SessionManager.Destroy(r.Context())

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
