package auth

import (
	"go-web-starter/cmd/web/views/auth"
	"go-web-starter/internal/forms"
	"go-web-starter/internal/forms/validator"
	"go-web-starter/internal/types"
	"net/http"

	"github.com/angelofallars/htmx-go"
)

func (ah *AuthHandler) ProfileViewHandler(w http.ResponseWriter, r *http.Request) {
	data := ah.handler.NewTemplateData(r)
	data.PageTitle = "Profile"

	updateUserForm := forms.UpdateUserNameAndImageForm{
		Name:  data.User.Name,
		Image: data.User.Image.String,
	}
	updatePasswordform := forms.UpdateAccountPasswordForm{}
	deleteAccountForm := forms.DeleteAccountForm{}

	auth.ProfileView(data, updateUserForm, updatePasswordform, deleteAccountForm).Render(r.Context(), w)
}

func (ah *AuthHandler) UpdateUserNameAndImageHandler(w http.ResponseWriter, r *http.Request) {
	var form forms.UpdateUserNameAndImageForm
	var data types.TemplateData

	err := ah.handler.DecodePostForm(r, &form)
	if err != nil {
		data = ah.handler.NewTemplateData(r)

		form.SetMessage("Invalid form data", forms.MessageTypeError)
		htmx.NewResponse().RenderTempl(r.Context(), w, auth.UpdateUserForm(data, form))
		return
	}

	// Validation
	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Image), "image", "This field cannot be blank")

	data = ah.handler.NewTemplateData(r)
	if !form.Valid() {
		htmx.NewResponse().RenderTempl(r.Context(), w, auth.UpdateUserForm(data, form))
		return
	}

	// Get the current user from context
	user := ah.handler.GetUser(r)

	_, err = ah.authService.UpdateUserNameAndImage(r.Context(), user.ID, form.Name, form.Image)
	if err != nil {
		form.SetMessage("Failed to update profile. Please try again.", forms.MessageTypeError)
	} else {
		form.SetMessage("Profile updated successfully!", forms.MessageTypeSuccess)
	}

	htmx.NewResponse().RenderTempl(r.Context(), w, auth.UpdateUserForm(data, form))
}

func (ah *AuthHandler) UpdateAccountPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var form forms.UpdateAccountPasswordForm
	var data types.TemplateData

	err := ah.handler.DecodePostForm(r, &form)
	if err != nil {
		data = ah.handler.NewTemplateData(r)

		form.SetMessage("Invalid form data", forms.MessageTypeError)
		htmx.NewResponse().RenderTempl(r.Context(), w, auth.ChangePasswordForm(data, form))

		return
	}

	// Validation
	form.CheckField(validator.NotBlank(form.CurrentPassword), "current_password", "Current password is required")
	form.CheckField(validator.NotBlank(form.NewPassword), "new_password", "New password is required")
	form.CheckField(validator.NotBlank(form.ConfirmPassword), "confirm_password", "Confirm password is required")

	data = ah.handler.NewTemplateData(r)

	if !form.Valid() {
		htmx.NewResponse().RenderTempl(r.Context(), w, auth.ChangePasswordForm(data, form))
		return
	}

	// Get the current user from context
	user := ah.handler.GetUser(r)

	err = ah.authService.UpdateAccountPassword(r.Context(), user.ID, form.CurrentPassword, form.NewPassword)
	if err != nil {
		if err.Error() == "invalid current password" {
			form.SetMessage("Current password is incorrect", forms.MessageTypeError)
		} else {
			form.SetMessage("Failed to update password. Please try again.", forms.MessageTypeError)
		}
	} else {
		form.SetMessage("Password updated successfully!", forms.MessageTypeSuccess)
	}

	htmx.NewResponse().RenderTempl(r.Context(), w, auth.ChangePasswordForm(data, form))
}

func (ah *AuthHandler) DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	var form forms.DeleteAccountForm
	var data types.TemplateData

	err := ah.handler.DecodePostForm(r, &form)
	if err != nil {
		data = ah.handler.NewTemplateData(r)
		form.SetMessage("Invalid form data", forms.MessageTypeError)
		htmx.NewResponse().RenderTempl(r.Context(), w, auth.DeleteAccountForm(data, form))
		return
	}

	// Validation
	form.CheckField(validator.NotBlank(form.Password), "password", "Password is required to delete your account")

	data = ah.handler.NewTemplateData(r)
	if !form.Valid() {
		htmx.NewResponse().RenderTempl(r.Context(), w, auth.DeleteAccountForm(data, form))
		return
	}

	// Get the current user from context
	user := ah.handler.GetUser(r)

	err = ah.authService.DeleteAccount(r.Context(), user.Email, form.Password)
	if err != nil {
		form.SetMessage("Failed to delete account. Please try again.", forms.MessageTypeError)
		htmx.NewResponse().RenderTempl(r.Context(), w, auth.DeleteAccountForm(data, form))
		return
	}

	// Clear session and redirect to home
	ah.handler.SessionManager.Destroy(r.Context())

	if htmx.IsHTMX(r) {
		htmx.NewResponse().Redirect("/").Write(w)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
