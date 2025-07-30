package auth

import (
	"go-htmx-sqlite/cmd/web/views/auth"
	"net/http"
)

type resetPasswordForm struct {
	Password             string `form:"password"`
	PasswordConfirmation string `form:"password-confirmation"`
	Token                string `form:"token"`
}

func (ah *AuthHandler) ResetPasswordPostHandler(w http.ResponseWriter, r *http.Request) {
	var resetPasswordForm resetPasswordForm
	err := ah.handler.DecodePostForm(r, &resetPasswordForm)
	if err != nil {
		ah.handler.ServerError(w, err)
		return
	}
	println(resetPasswordForm.Password, resetPasswordForm.Token)
	// validate

	/*
		AuthService: verify the token,reset the password, delete the token from database
	*/
	user, err := ah.authService.ResetPassword(r.Context(), resetPasswordForm.Token, resetPasswordForm.Password)
	if err != nil {
		ah.handler.ServerError(w, err)
		return
	}

	// notify the user by mail
	println("user password has been updated", user.Email)

	// redirect to login page with session 'flash' message
	ah.handler.SessionManager.Put(r.Context(), "flash", "Password reset successfully")

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (ah *AuthHandler) ResetPasswordView(w http.ResponseWriter, r *http.Request) {
	data := ah.handler.NewTemplateData(r)
	data.PageTitle = "Reset Password"

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
	auth.ResetPasswordView(data).Render(r.Context(), w)
}
