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

func (ah *AuthHandler) ResetPasswordPostHandler(w http.ResponseWriter, r *http.Request) {
	var form forms.ResetPasswordForm

	err := ah.handler.DecodePostForm(r, &form)
	if err != nil {
		htmx.NewResponse().RenderTempl(
			r.Context(),
			w,
			components.FlashMessage("invalid form data", components.FlashError),
		)
		return
	}

	// Validation
	form.CheckField(validator.NotBlank(form.Token), "token", "token is required")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	if !form.Valid() {
		data := ah.handler.NewTemplateData(r)

		htmx.NewResponse().RenderTempl(r.Context(), w, auth.PasswordResetForm(data, form))

		return
	}

	// AuthService: verify the token, reset the password, delete the token from database
	user, err := ah.authService.ResetPassword(r.Context(), form.Token, form.Password)
	if err != nil {
		htmx.NewResponse().RenderTempl(
			r.Context(),
			w,
			components.FlashMessage("Invalid or expired reset token. Please request a new password reset.", components.FlashError),
		)
		return
	}

	// Notify the user by mail
	err = ah.handler.Mailer.Send(user.Email, "reset_password_confirmation.tmpl", nil)
	if err != nil {
		ah.handler.Logger.PrintError(err, nil)
	}

	htmx.NewResponse().RenderTempl(
		r.Context(),
		w,
		components.FlashMessage("Password reset successfully! You can now log in with your new password.", components.FlashSuccess),
	)
}

func (ah *AuthHandler) ResetPasswordView(w http.ResponseWriter, r *http.Request) {
	data := ah.handler.NewTemplateData(r)
	data.PageTitle = "Reset Password"
	form := forms.ResetPasswordForm{}

	// get the token from query ?token=
	plainTextToken := r.URL.Query().Get("token")

	// validate the token format - should be 26 characters (base32 encoded 16 bytes)
	if len(plainTextToken) != 26 {
		err := fmt.Errorf("invalid token format: expected 26 characters, got %d", len(plainTextToken))
		ah.handler.ServerError(w, err)
		return
	}

	// compare the token with the hashed one in the database
	_, err := ah.authService.GetValidTokenUser(r.Context(), plainTextToken)
	if err != nil {
		ah.handler.ServerError(w, err)
		return
	}

	data.Meta = map[string]string{
		"token": plainTextToken,
	}

	// get the user to the reset-password page with the token in an input named "token"
	auth.ResetPasswordView(data, form).Render(r.Context(), w)
}
