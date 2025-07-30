package auth

import (
	"go-htmx-sqlite/cmd/web/views/auth"
	"log"
	"net/http"
	"net/url"
)

type userLoginForm struct {
	Email    string `form:"email"`
	Password string `form:"password"`
}

func (ah *AuthHandler) LoginPostHandler(w http.ResponseWriter, r *http.Request) {
	var loginForm userLoginForm

	err := ah.handler.DecodePostForm(r, &loginForm)
	if err != nil {
		log.Panic(err)
	}

	// validation
	// handle validation errors
	// authenticate: check the user and account exists
	user, err := ah.authService.Login(r.Context(), loginForm.Email, loginForm.Password)
	if err != nil {
		log.Println(err)
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
