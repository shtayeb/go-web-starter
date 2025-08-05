package auth

import (
	"go-htmx-sqlite/cmd/web/components"
	"go-htmx-sqlite/cmd/web/views/auth"
	"go-htmx-sqlite/internal/types"
	"go-htmx-sqlite/internal/validator"
	"net/http"

	"github.com/angelofallars/htmx-go"
)

func (ah *AuthHandler) ProfileViewHandler(w http.ResponseWriter, r *http.Request) {
	data := ah.handler.NewTemplateData(r)
	data.PageTitle = "Profile"

	updateUserForm := types.UpdateUserNameAndImageForm{
		Name:  data.User.Name,
		Image: data.User.Image.String,
	}
	updatePasswordform := types.UpdateAccountPasswordForm{}

	auth.ProfileView(data, updateUserForm, updatePasswordform).Render(r.Context(), w)
}

func (ah *AuthHandler) UpdateUserNameAndImageHandler(w http.ResponseWriter, r *http.Request) {
	var form types.UpdateUserNameAndImageForm

	err := ah.handler.DecodePostForm(r, &form)
	if err != nil {
		htmx.NewResponse().RenderTempl(r.Context(), w,
			components.FlashMessage("invalid form data"),
		)
		return
	}

	// Validation
	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Image), "image", "This field cannot be blank")

	if !form.Valid() {
		data := ah.handler.NewTemplateData(r)

		htmx.NewResponse().RenderTempl(r.Context(), w, auth.UpdateUserForm(data, form))
		return
	}

	// Get the current user from context
	user := ah.handler.GetUser(r)

	_, err = ah.authService.UpdateUserNameAndImage(r.Context(), user.ID, form.Name, form.Image)
	if err != nil {
		htmx.NewResponse().RenderTempl(r.Context(), w,
			components.FlashMessage("Failed to update profile. Please try again."),
		)
		return
	}

	htmx.NewResponse().RenderTempl(r.Context(), w,
		components.FlashMessage("Profile updated successfully!"),
	)
}

func (ah *AuthHandler) UpdateAccountPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var form types.UpdateAccountPasswordForm

	err := ah.handler.DecodePostForm(r, &form)
	if err != nil {
		htmx.NewResponse().RenderTempl(r.Context(), w,
			components.FlashMessage("Invalid form data"),
		)
		return
	}

	// Validation
	form.CheckField(validator.NotBlank(form.CurrentPassword), "current_password", "Current password is required")
	form.CheckField(validator.NotBlank(form.NewPassword), "new_password", "New password is required")
	form.CheckField(validator.NotBlank(form.ConfirmPassword), "confirm_password", "Confirm password is required")

	if !form.Valid() {
		data := ah.handler.NewTemplateData(r)

		htmx.NewResponse().RenderTempl(r.Context(), w, auth.ChangePasswordForm(data, form))
		return
	}

	// Get the current user from context
	user := ah.handler.GetUser(r)

	err = ah.authService.UpdateAccountPassword(r.Context(), user.ID, form.CurrentPassword, form.NewPassword)
	if err != nil {
		if err.Error() == "invalid current password" {
			htmx.NewResponse().RenderTempl(r.Context(), w,
				components.FlashMessage("Current password is incorrect"),
			)
			return
		}
		htmx.NewResponse().RenderTempl(r.Context(), w,
			components.FlashMessage("Failed to update password. Please try again."),
		)
		return
	}

	htmx.NewResponse().RenderTempl(r.Context(), w,
		components.FlashMessage("Password updated successfully!"),
	)
}
