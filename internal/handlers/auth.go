package handlers

import (
	"go-htmx-sqlite/cmd/web/views/auth"
	"net/http"
)

func (h *Handlers) LoginViewHandler(w http.ResponseWriter, r *http.Request) {
	// check user shouldnt be logged in
	auth.LoginView().Render(r.Context(), w)
}

func (h *Handlers) SignUpViewHandler(w http.ResponseWriter, r *http.Request) {
	// check user shouldnt be logged in
	auth.SignUpView().Render(r.Context(), w)
}
