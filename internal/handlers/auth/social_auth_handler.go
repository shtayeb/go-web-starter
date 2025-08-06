package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/markbates/goth/gothic"
)

func (ah *AuthHandler) SocialAuthHandler(w http.ResponseWriter, r *http.Request) {
	// try to get the user without re-authenticating
	if gothUser, err := gothic.CompleteUserAuth(w, r); err == nil {
		// Convert gothUser to JSON and write to response
		userJSON, err := json.Marshal(gothUser)
		if err != nil {
			http.Error(w, "Error encoding user data", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(userJSON)
	} else {
		gothic.BeginAuthHandler(w, r)
	}
}

func (ah *AuthHandler) SocialAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// This is after the social login
	// provider := chi.URLParam(r, "provider")

	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	w.Write([]byte("User from the callback: " + user.Email))

	// handle redirects and sessions in here
}
