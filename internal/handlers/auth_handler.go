package handlers

import (
	"go-htmx-sqlite/cmd/web/views/auth"
	"go-htmx-sqlite/internal/queries"
	"log"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (h *Handlers) LogoutPostHandler(w http.ResponseWriter, r *http.Request) {
	err := h.SessionManager.RenewToken(r.Context())
	if err != nil {
		h.serverError(w, err)
		return
	}

	h.SessionManager.Remove(r.Context(), "authenticatedUserID")
	h.SessionManager.Remove(r.Context(), "user")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) ResetPasswordView(w http.ResponseWriter, r *http.Request) {
	data := h.newTemplateData(r)
	data.PageTitle = "Reset Password"

	// Get the token from url params
	auth.ResetPasswordView(data).Render(r.Context(), w)
}

func (h *Handlers) ForgotPasswordView(w http.ResponseWriter, r *http.Request) {
	data := h.newTemplateData(r)
	data.PageTitle = "Forgot Password"

	auth.ForgotPasswordView(data).Render(r.Context(), w)
}

func (h *Handlers) ForgotPasswordPostHanlder(w http.ResponseWriter, r *http.Request) {
	// handle form and its validation
	type ForgotPasswordForm struct {
		Email string `form:"email"`
	}

	var forgotPasswordForm ForgotPasswordForm

	err := h.decodePostForm(r, &forgotPasswordForm)
	if err != nil {
		log.Panic(err)
	}

	// get the user by email
	user, err := h.DB.GetUserByEmail(r.Context(), forgotPasswordForm.Email)
	if err != nil {
		return
	}

	println(user.Email)
	// handle the errors in the view

	// create token with ttl of 15min
	// send the reset email with the token for the user
}

func (h *Handlers) LoginViewHandler(w http.ResponseWriter, r *http.Request) {
	data := h.newTemplateData(r)
	data.PageTitle = "Login"

	auth.LoginView(data).Render(r.Context(), w)
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

	err := h.decodePostForm(r, &loginForm)
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

	// fmt.Printf("\n %#v \n %#v", loginForm.Password, account.Password)
	if !checkPasswordHash(account.Password, loginForm.Password) {
		// invalid password - handle errors in login page
		w.Write([]byte("not match"))
		return
	}

	// session manager
	err = h.SessionManager.RenewToken(r.Context())
	if err != nil {
		return
	}

	h.SessionManager.Put(r.Context(), "authenticatedUserID", user.ID)

	// get the next=? query string if exists. 1 - redirect to it. 2 -  or redirect to home after login
	redirectURL := "/dashboard"
	refererUrl, _ := url.Parse(r.Referer())
	path := refererUrl.Query().Get("next")
	if path != "" {
		redirectURL = path
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (h *Handlers) SignUpViewHandler(w http.ResponseWriter, r *http.Request) {
	// check user shouldnt be logged in
	data := h.newTemplateData(r)
	data.PageTitle = "Sign Up"

	auth.SignUpView(data).Render(r.Context(), w)
}

type userSignupForm struct {
	Name                 string `form:"name"`
	Email                string `form:"email"`
	Password             string `form:"password"`
	PasswordConfirmation string `form:"password-confirmation"`
}

func (h *Handlers) SignUpPostHandler(w http.ResponseWriter, r *http.Request) {
	var signUpForm userSignupForm

	err := h.decodePostForm(r, &signUpForm)
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
	_, err = h.DB.CreateAccount(r.Context(), queries.CreateAccountParams{
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

	// send the user a message to verify the user's email address account
	// TODO:jwtToken with a ttl of 6 hour
	jwtToken := "hello"
	data := map[string]any{
		"activationToken": jwtToken,
		"userID":          user.ID,
	}

	// TODO:Send this to a background job handler, where it can be retried
	err = h.Mailer.Send(user.Email, "user_welcome.tmpl", data)
	if err != nil {
		h.Logger.PrintError(err, nil)
	}

	// add message to the session manager and display it to the user
	h.SessionManager.Put(r.Context(), "flash", "Your account was created successfully!")

	// redirect to the login page
	// fmt.Printf("%#v \n %#v", user, account)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
