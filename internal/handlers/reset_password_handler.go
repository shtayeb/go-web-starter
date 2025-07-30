package handlers

import (
	"crypto/sha256"
	"go-htmx-sqlite/cmd/web/views/auth"
	"go-htmx-sqlite/internal/queries"
	"net/http"
	"time"
)

func (h *Handlers) ResetPasswordPostHandler(w http.ResponseWriter, r *http.Request) {
	// form

	// validate

	// verify the token

	// reset the password

	// delete the token from database

	// notify the user by mail

	// redirect to login page
}

func (h *Handlers) ResetPasswordView(w http.ResponseWriter, r *http.Request) {
	data := h.newTemplateData(r)
	data.PageTitle = "Reset Password"

	// get the token from query ?token=
	plainTextToken := r.URL.Query().Get("token")

	// validate the token -> should be 26 byte length and its

	// hash the plainText token
	tokenHash := sha256.Sum256([]byte(plainTextToken))

	println(plainTextToken, tokenHash[:], time.Now().String())

	// compare the token with the hashed one in the database
	_, err := h.DB.GetUserByToken(r.Context(), queries.GetUserByTokenParams{
		Hash:   tokenHash[:],
		Scope:  ScopePasswordReset,
		Expiry: time.Now(),
	})
	if err != nil {
		// if not match:
		//	- return with error
		println(err.Error())
		return
	}

	data.Meta = map[string]string{
		"token": plainTextToken,
	}

	// get the user to the reset-password page with the token in an input named "token"
	auth.ResetPasswordView(data).Render(r.Context(), w)
}
