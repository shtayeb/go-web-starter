package auth

import (
	"go-htmx-sqlite/cmd/web/views/auth"
	"log"
	"net/http"
	"net/url"

	"golang.org/x/crypto/bcrypt"
)

type userLoginForm struct {
	Email    string `form:"email"`
	Password string `form:"password"`
}

func checkPasswordHash(hashedPassword, plainTextPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainTextPassword))
	return err == nil
}

func (ah *AuthHandler) LoginPostHandler(w http.ResponseWriter, r *http.Request) {
	var loginForm userLoginForm

	err := ah.handler.DecodePostForm(r, &loginForm)
	if err != nil {
		log.Panic(err)
	}

	// validation
	// handle validation errors

	// authenticate: check the email exists
	user, err := ah.handler.DB.GetUserByEmail(r.Context(), loginForm.Email)
	if err != nil {
		// handle error
		log.Panic(err)
	}

	account, err := ah.handler.DB.GetAccountByUserId(r.Context(), user.ID)
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
	err = ah.handler.SessionManager.RenewToken(r.Context())
	if err != nil {
		return
	}

	ah.handler.SessionManager.Put(r.Context(), "authenticatedUserID", user.ID)

	// get the next=? query string if exists. 1 - redirect to it. 2 -  or redirect to home after login
	redirectURL := "/dashboard"
	refererUrl, _ := url.Parse(r.Referer())
	path := refererUrl.Query().Get("next")
	if path != "" {
		redirectURL = path
	}

	ah.handler.SessionManager.Put(r.Context(), "flash", "reset link have been sent to your email")

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (ah *AuthHandler) LoginViewHandler(w http.ResponseWriter, r *http.Request) {
	data := ah.handler.NewTemplateData(r)
	data.PageTitle = "Login"

	auth.LoginView(data).Render(r.Context(), w)
}
