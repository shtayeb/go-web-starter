package auth

import (
	"errors"
	"go-htmx-sqlite/cmd/web/views/auth"
	"net/http"
)

func (ah *AuthHandler) ProfileViewHandler(w http.ResponseWriter, r *http.Request) {
	data := ah.handler.NewTemplateData(r)
	data.PageTitle = "Profile"

	auth.ProfileView(data).Render(r.Context(), w)
}

type UpdateUserNameAndImageForm struct {
	Name  string `form:"name"`
	Image string `form:"image"`
}

func (ah *AuthHandler) UpdateUserNameAndImageHandler(w http.ResponseWriter, r *http.Request) {
	var form UpdateUserNameAndImageForm

	err := ah.handler.DecodePostForm(r, &form)
	if err != nil {
		ah.handler.ServerError(w, err)
		return
	}
	// validation and handle validation errors

	// get the current user from context
	user := ah.handler.GetUser(r)

	_, err = ah.authService.UpdateUserNameAndImage(r.Context(), user.ID, form.Name, form.Image)
	if err != nil {
		ah.handler.ServerError(w, err)
		return
	}

	ah.handler.SessionManager.Put(r.Context(), "flash", "Profile updated successfully")
	// handle the flash with htmx

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

type UpdateAccountPasswordForm struct {
	CurrentPassword string `form:"current_password"`
	NewPassword     string `form:"new_password"`
	ConfirmPassword string `form:"password_confirmation"`
}

func (ah *AuthHandler) UpdateAccountPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var form UpdateAccountPasswordForm

	err := ah.handler.DecodePostForm(r, &form)
	if err != nil {
		ah.handler.ServerError(w, err)
		return
	}

	// validation and handle validation errors
	if form.NewPassword != form.ConfirmPassword {
		ah.handler.ServerError(w, errors.New("passwords do not match"))
		return
	}

	// get the current user from context
	user := ah.handler.GetUser(r)

	err = ah.authService.UpdateAccountPassword(r.Context(), user.ID, form.CurrentPassword, form.NewPassword)
	if err != nil {
		ah.handler.ServerError(w, err)
		return
	}

	ah.handler.SessionManager.Put(r.Context(), "flash", "Profile updated successfully")
	// handle the flash with htmx

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
