package auth

import (
	"go-htmx-sqlite/cmd/web/views/auth"
	"net/http"
)

func (ah *AuthHandler) ProfileViewHandler(w http.ResponseWriter, r *http.Request) {
	data := ah.handler.NewTemplateData(r)
	data.PageTitle = "Profile"

	auth.ProfileView(data).Render(r.Context(), w)
}
