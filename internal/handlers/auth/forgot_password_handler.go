package auth

import (
	"go-htmx-sqlite/cmd/web/views/auth"
	"log"
	"net/http"
)

func (ah *AuthHandler) ForgotPasswordView(w http.ResponseWriter, r *http.Request) {
	data := ah.handler.NewTemplateData(r)
	data.PageTitle = "Forgot Password"

	auth.ForgotPasswordView(data).Render(r.Context(), w)
}

func (ah *AuthHandler) ForgotPasswordPostHanlder(w http.ResponseWriter, r *http.Request) {
	// handle form and its validation
	type ForgotPasswordForm struct {
		Email string `form:"email"`
	}

	var forgotPasswordForm ForgotPasswordForm

	err := ah.handler.DecodePostForm(r, &forgotPasswordForm)
	if err != nil {
		log.Panic(err)
	}

	passwordResetLink, err := ah.authService.GetPasswordResetLink(r.Context(), forgotPasswordForm.Email)
	if err != nil {
		ah.handler.Logger.PrintError(err, nil)
		return
	}

	// send the reset email with the token for the user
	data := map[string]any{
		"passwordResetLink": passwordResetLink,
	}
	err = ah.handler.Mailer.Send(forgotPasswordForm.Email, "reset_password.tmpl", data)
	if err != nil {
		ah.handler.Logger.PrintError(err, nil)
	}

	// set a flash message in the session manager
	ah.handler.SessionManager.Put(r.Context(), "flash", "Link sent to your email address")

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
