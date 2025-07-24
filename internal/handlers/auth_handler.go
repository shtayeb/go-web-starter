package handlers

import (
	"fmt"
	"go-htmx-sqlite/cmd/web/views/auth"
	"go-htmx-sqlite/internal/queries"
	"log"
	"net/http"
	"time"

	"github.com/go-playground/form/v4"
	"golang.org/x/crypto/bcrypt"
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

type userLoginForm struct {
	Email    string `form:"email"`
	Password string `form:"password"`
}

func hashPassword(plainTextPassword string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), 14)
	return string(bytes), err
}

func checkPasswordHash(hashedPassword, plainTextPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainTextPassword))
	return err == nil
}

func (h *Handlers) LoginPostHandler(w http.ResponseWriter, r *http.Request) {
	var loginForm userLoginForm

	err := r.ParseForm()
	if err != nil {
		log.Panic(err)
	}

	decoder := form.NewDecoder()
	err = decoder.Decode(&loginForm, r.PostForm)
	if err != nil {
		log.Panic(err)
	}

	// validation
	// handle validation errors

	// authenticate: check the email exists
	user, err := h.DB.GetUserByEmail(r.Context(), loginForm.Email)
	if err != nil {
		// handle error
		log.Panic(err)
	}

	account, err := h.DB.GetAccountByUserId(r.Context(), user.ID)
	if err != nil {
		// handle error
		log.Panic(err)
	}

	fmt.Printf("\n %#v \n %#v", loginForm.Password, account.Password)

	if !checkPasswordHash(account.Password, loginForm.Password) {
		// invalid password - handle errors in login page
		w.Write([]byte("not match"))
		return
	}

	// session manager

	// get the next=? query string if exists. 1 - redirect to it. 2 -  or redirect to home after login
	redirectURL := "/dashboard"
	if next := r.URL.Query().Get("next"); next != "" {
		// validate the redirect URL to prevent open redirect vulnerabilities
		redirectURL = next
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
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

	// insert into the users table
	user, err := h.DB.CreateUser(r.Context(), queries.CreateUserParams{
		Name:      signUpForm.Name,
		Email:     signUpForm.Email,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		h.Logger.PrintError(err, map[string]string{
			"request_method": r.Method,
			"request_url":    r.URL.String(),
		})

		panic(err)
		// handle the error in the frontend give user a error message
	}

	// hash the password
	hashedPassword, err := hashPassword(signUpForm.Password)
	if err != nil {
		return
		// handle error in the view
	}

	// insert into the accounts table - password is in the accounts table
	account, err := h.DB.CreateAccount(r.Context(), queries.CreateAccountParams{
		UserID:    user.ID,
		AccountID: user.Name,
		Password:  string(hashedPassword),
	})
	// handle database errors
	if err != nil {
		h.Logger.PrintError(err, map[string]string{
			"request_method": r.Method,
			"request_url":    r.URL.String(),
		})

		// handle the error in the frontend give user an error message
	}

	// add message to the session manager and display it to the user

	// redirect to the login page
	fmt.Printf("%#v \n %#v", user, account)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
