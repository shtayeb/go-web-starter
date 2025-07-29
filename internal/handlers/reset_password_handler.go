package handlers

import (
	"go-htmx-sqlite/cmd/web/views/auth"
	"net/http"
)

func (h *Handlers) ResetPasswordView(w http.ResponseWriter, r *http.Request) {
	data := h.newTemplateData(r)
	data.PageTitle = "Reset Password"

	// Get the token from url params
	auth.ResetPasswordView(data).Render(r.Context(), w)
}
