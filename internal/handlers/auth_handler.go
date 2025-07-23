package handlers

import (
	"fmt"
	"go-htmx-sqlite/cmd/web/views/auth"
	"log"
	"net/http"

	"github.com/go-playground/form/v4"
)

func (h *Handlers) ResetPasswordView(w http.ResponseWriter, r *http.Request) {
	// check user shouldnt be logged in

	// Get the token from url params
	auth.ResetPasswordView().Render(r.Context(), w)
}

func (h *Handlers) ForgotPasswordView(w http.ResponseWriter, r *http.Request) {
	// check user shouldnt be logged in
	auth.ForgotPasswordView().Render(r.Context(), w)
}

func (h *Handlers) LoginViewHandler(w http.ResponseWriter, r *http.Request) {
	// check user shouldnt be logged in
	auth.LoginView().Render(r.Context(), w)
}

func (h *Handlers) SignUpViewHandler(w http.ResponseWriter, r *http.Request) {
	// check user shouldnt be logged in
	auth.SignUpView().Render(r.Context(), w)
}

type userSignupForm struct {
	Name                 string `form:"name"`
	Email                string `form:"email"`
	Password             string `form:"password"`
	PasswordConfirmation string `form:"password-confirmation"`
}

func (h *Handlers) SignUpPostHandler(w http.ResponseWriter, r *http.Request) {
	var signUpForm userSignupForm

	err := r.ParseForm()
	if err != nil {
		log.Panic(err)
	}

	decoder := form.NewDecoder()

	err = decoder.Decode(&signUpForm, r.PostForm)
	if err != nil {
		log.Panic(err)
	}

	// validate the form
	// Handle if not valid - return the validation errors

	// handle valid

	// hash the password
	// TODO: insert into the users table

	// insert into the accounts table - password is in the accounts table
	// handle database errors

	// add message to the session manager and display it to the user

	// redirect to the login page

	fmt.Printf("%#v\n", signUpForm)
}
