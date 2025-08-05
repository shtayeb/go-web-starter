package auth

import (
	"go-htmx-sqlite/cmd/web/components"
	"go-htmx-sqlite/cmd/web/views/auth"
	"go-htmx-sqlite/internal/types"
	"go-htmx-sqlite/internal/validator"
	"net/http"

	"github.com/angelofallars/htmx-go"
)

func (ah *AuthHandler) ResetPasswordPostHandler(w http.ResponseWriter, r *http.Request) {
	var form types.ResetPasswordForm

	err := ah.handler.DecodePostForm(r, &form)
	if err != nil {
		htmx.NewResponse().RenderTempl(
			r.Context(),
			w,
			components.FlashMessage("invalid form data"),
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
			components.FlashMessage("Invalid or expired reset token. Please request a new password reset."),
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
		components.FlashMessage("Password reset successfully! You can now log in with your new password."),
	)
}

func (ah *AuthHandler) ResetPasswordView(w http.ResponseWriter, r *http.Request) {
	data := ah.handler.NewTemplateData(r)
	data.PageTitle = "Reset Password"
	form := types.ResetPasswordForm{}

	// get the token from query ?token=
	plainTextToken := r.URL.Query().Get("token")
	// validate the token -> should be 26 byte length and its

	// compare the token with the hashed one in the database
	_, err := ah.authService.GetValidTokenUser(r.Context(), plainTextToken)
	if err != nil {
		// if not match:
		//	- return with error
		ah.handler.ServerError(w, err)
		return
	}

	data.Meta = map[string]string{
		"token": plainTextToken,
	}

	// get the user to the reset-password page with the token in an input named "token"
	auth.ResetPasswordView(data, form).Render(r.Context(), w)
}
