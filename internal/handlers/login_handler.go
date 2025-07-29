package handlers

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

	h.SessionManager.Put(r.Context(), "flash", "reset link have been sent to your email")

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (h *Handlers) LoginViewHandler(w http.ResponseWriter, r *http.Request) {
	data := h.newTemplateData(r)
	data.PageTitle = "Login"

	auth.LoginView(data).Render(r.Context(), w)
}
