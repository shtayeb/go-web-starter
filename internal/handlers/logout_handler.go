package handlers

import (
	"net/http"
)

func (h *Handlers) LogoutPostHandler(w http.ResponseWriter, r *http.Request) {
	err := h.SessionManager.RenewToken(r.Context())
	if err != nil {
		h.serverError(w, err)
		return
	}

	h.SessionManager.Remove(r.Context(), "authenticatedUserID")
	h.SessionManager.Remove(r.Context(), "user")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
