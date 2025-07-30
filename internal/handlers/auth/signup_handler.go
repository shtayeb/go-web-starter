package auth

import (
	"go-htmx-sqlite/cmd/web/views/auth"
	"log"
	"net/http"
)

type userSignupForm struct {
	Name                 string `form:"name"`
	Email                string `form:"email"`
	Password             string `form:"password"`
	PasswordConfirmation string `form:"password-confirmation"`
}

func (ah *AuthHandler) SignUpPostHandler(w http.ResponseWriter, r *http.Request) {
	var signUpForm userSignupForm

	err := ah.handler.DecodePostForm(r, &signUpForm)
	if err != nil {
		log.Panic(err)
	}

	// validate the form
	// Handle if not valid - return the validation errors
	// handle valid

	// insert into the users table - with DB transaction
	user, err := ah.authService.SignUp(r.Context(), signUpForm.Name, signUpForm.Email, signUpForm.Password)
	// handle database errors
	if err != nil {
		ah.handler.Logger.PrintError(err, map[string]string{
			"request_method": r.Method,
			"request_url":    r.URL.String(),
		})

		// handle the error in the frontend give user an error message
		ah.handler.ServerError(w, err)
		return
	}

	// send the user a message to verify the user's email address account
	// TODO:token with a ttl of 6 hour
	activationLink := "hello"
	data := map[string]any{
		"activationLink": activationLink,
		"userID":         user.ID,
	}
	// TODO:Send this to a background job handler, where it can be retried
	err = ah.handler.Mailer.Send(user.Email, "user_welcome.tmpl", data)
	if err != nil {
		ah.handler.Logger.PrintError(err, nil)
	}

	// add message to the session manager and display it to the user
	ah.handler.SessionManager.Put(r.Context(), "flash", "Your account was created successfully!")

	// redirect to the login page
	// fmt.Printf("%#v \n %#v", user, account)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (ah *AuthHandler) SignUpViewHandler(w http.ResponseWriter, r *http.Request) {
	// check user shouldnt be logged in
	data := ah.handler.NewTemplateData(r)
	data.PageTitle = "Sign Up"

	auth.SignUpView(data).Render(r.Context(), w)
}
