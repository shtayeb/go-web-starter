package handlers

import (
	"go-htmx-sqlite/cmd/web/views/auth"
	"go-htmx-sqlite/internal/queries"
	"log"
	"net/http"
	"time"
)

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

func (h *Handlers) SignUpViewHandler(w http.ResponseWriter, r *http.Request) {
	// check user shouldnt be logged in
	data := h.newTemplateData(r)
	data.PageTitle = "Sign Up"

	auth.SignUpView(data).Render(r.Context(), w)
}
