package auth

import (
	"go-htmx-sqlite/cmd/web/components"
	"go-htmx-sqlite/cmd/web/views/auth"
	"go-htmx-sqlite/internal/types"
	"go-htmx-sqlite/internal/validator"
	"net/http"

	"github.com/angelofallars/htmx-go"
)

func (ah *AuthHandler) LoginPostHandler(w http.ResponseWriter, r *http.Request) {
	var form types.UserLoginForm

	err := ah.handler.DecodePostForm(r, &form)
	if err != nil {
		htmx.NewResponse().RenderTempl(r.Context(), w, components.FlashMessage("Invalid form data"))
		return
	}

	// Validation
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		// handle with htmx
		htmx.NewResponse().RenderTempl(r.Context(), w, components.FlashMessage("Something went wrong!"))
		return
	}

	// Authenticate: check the user and account exists
	user, err := ah.authService.Login(r.Context(), form.Email, form.Password)
	if err != nil {
		htmx.NewResponse().RenderTempl(r.Context(), w, components.FlashMessage("Invalid email or password"))
		return
	}

	// Session manager - renew token AFTER successful authentication to prevent session fixation
	err = ah.handler.SessionManager.RenewToken(r.Context())
	if err != nil {
		htmx.NewResponse().RenderTempl(r.Context(), w, components.FlashMessage("Session error occurred"))
		return
	}

	ah.handler.SessionManager.Put(r.Context(), "authenticatedUserID", user.ID)

	// Get the next=? query string if exists. 1 - redirect to it. 2 - or redirect to home after login
	redirectURL := "/dashboard"
	nextPath := r.URL.Query().Get("next")

	if nextPath != "" && IsValidRedirectPath(nextPath) {
		redirectURL = nextPath
	}

	if htmx.IsHTMX(r) {
		htmx.NewResponse().Redirect(redirectURL).Write(w)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (ah *AuthHandler) LoginViewHandler(w http.ResponseWriter, r *http.Request) {
	data := ah.handler.NewTemplateData(r)
	data.PageTitle = "Login"

	form := types.UserLoginForm{}

	auth.LoginView(data, form).Render(r.Context(), w)
}
