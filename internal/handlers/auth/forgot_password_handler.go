package auth

import (
	"fmt"
	"go-htmx-sqlite/cmd/web/components"
	"go-htmx-sqlite/cmd/web/views/auth"
	"go-htmx-sqlite/internal/forms"
	"go-htmx-sqlite/internal/forms/validator"
	"net/http"

	"github.com/angelofallars/htmx-go"
)

func (ah *AuthHandler) ForgotPasswordView(w http.ResponseWriter, r *http.Request) {
	data := ah.handler.NewTemplateData(r)
	data.PageTitle = "Forgot Password"
	form := forms.ForgotPasswordForm{}

	auth.ForgotPasswordView(data, form).Render(r.Context(), w)
}

func (ah *AuthHandler) ForgotPasswordPostHanlder(w http.ResponseWriter, r *http.Request) {
	var form forms.ForgotPasswordForm

	err := ah.handler.DecodePostForm(r, &form)
	if err != nil {
		htmx.NewResponse().RenderTempl(r.Context(), w,
			components.FlashMessage("invalid form data", components.FlashError),
		)
		return
	}

	// Validation
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")

	if !form.Valid() {
		// handle with htmx
		data := ah.handler.NewTemplateData(r)
		htmx.NewResponse().RenderTempl(r.Context(), w, auth.ForgotPasswordForm(data, form))
		return
	}

	// Construct base URL from request
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, r.Host)

	passwordResetLink, err := ah.authService.GetPasswordResetLink(r.Context(), form.Email, baseURL)
	if err != nil {
		ah.handler.Logger.PrintError(err, nil)
		// Don't reveal if email exists or not for security
		htmx.NewResponse().RenderTempl(r.Context(), w,
			components.FlashMessage("Can not send you a password reset link. Please try again later!", components.FlashError),
		)
		return
	}

	// Send the reset email with the token for the user
	data := map[string]any{
		"passwordResetLink": passwordResetLink,
	}
	err = ah.handler.Mailer.Send(form.Email, "reset_password.tmpl", data)
	if err != nil {
		ah.handler.Logger.PrintError(err, nil)
	}

	htmx.NewResponse().RenderTempl(r.Context(), w,
		components.FlashMessage("If an account with this email exists, a password reset will be sent to it.", components.FlashInfo),
	)
}
