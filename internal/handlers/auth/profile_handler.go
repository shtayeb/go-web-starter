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
