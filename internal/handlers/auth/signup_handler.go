package auth

import (
	"go-htmx-sqlite/cmd/web/components"
	"go-htmx-sqlite/cmd/web/views/auth"
	"go-htmx-sqlite/internal/forms"
	"go-htmx-sqlite/internal/forms/validator"
	"net/http"

	"github.com/angelofallars/htmx-go"
)

func (ah *AuthHandler) SignUpPostHandler(w http.ResponseWriter, r *http.Request) {
	var form forms.UserSignUpForm

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
	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")
	form.CheckField(validator.NotBlank(form.ConfirmPassword), "confirm_password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.ConfirmPassword, 8), "confirm_password", "This field must be at least 8 characters long")
	form.CheckField(validator.Equals(form.Password, form.ConfirmPassword), "confirm_password", "Passwords do not match")

	if !form.Valid() {
		// handle with htmx
		data := ah.handler.NewTemplateData(r)
		htmx.NewResponse().RenderTempl(r.Context(), w, auth.SignUpForm(data, form))
		return
	}

	// Insert into the users table - with DB transaction
	user, err := ah.authService.SignUp(r.Context(), form.Name, form.Email, form.Password)
	if err != nil {
		ah.handler.Logger.PrintError(err, map[string]string{
			"request_method": r.Method,
			"request_url":    r.URL.String(),
		})

		// Handle common database errors
		htmx.NewResponse().RenderTempl(r.Context(), w, components.FlashMessage("Something went with your registration. please try again", components.FlashError))
		return
	}

	// Send the user a message to verify the user's email address account
	// TODO: token with a ttl of 6 hour
	activationLink := "http://localhost:8080/activate?token="
	data := map[string]any{
		"activationLink": activationLink,
		"userID":         user.ID,
	}
	// TODO: Send this to a background job handler, where it can be retried
	err = ah.handler.Mailer.Send(user.Email, "user_welcome.tmpl", data)
	if err != nil {
		ah.handler.Logger.PrintError(err, nil)
	}

	// add message to the session manager and display it to the user
	ah.handler.SessionManager.Put(r.Context(), "flash", "Your account was created successfully!")

	// NOTE: could create a helper for this
	if htmx.IsHTMX(r) {
		htmx.NewResponse().Redirect("/login").Write(w)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (ah *AuthHandler) SignUpViewHandler(w http.ResponseWriter, r *http.Request) {
	// check user shouldnt be logged in
	data := ah.handler.NewTemplateData(r)
	data.PageTitle = "Sign Up"

	form := forms.UserSignUpForm{}

	auth.SignUpView(data, form).Render(r.Context(), w)
}
