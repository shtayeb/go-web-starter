package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"go-htmx-sqlite/cmd/web/views/auth"
	"go-htmx-sqlite/internal/queries"
	"log"
	"net/http"
	"time"
)

const ScopeActivation = "activation"
const ScopeAuthentication = "authentication"
const ScopePasswordReset = "password-reset"

func generateToken(userID int64, ttl time.Duration, scope string) ([]byte, string, error) {
	// Create a Token instance containing the user ID, expiry, and scope information.
	// token := &queries.Token{
	// 	UserID: userID,
	// 	Expiry: time.Now().Add(ttl),
	// 	Scope:  scope,
	// }

	// Initialize a zero-valued byte slice with a length of 16 bytes.
	randomBytes := make([]byte, 16)

	// Use the Read() function from the crypto/rand package to fill the byte slice with
	// random bytes from your operating system's CSPRNG.
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, "", err
	}

	// Encode the byte slice to a base-32-encoded string and assign it to the token
	// Plaintext field. This will be the token string that we send to the user in their
	// welcome email. They will look similar to this:
	//
	// Y3QMGX3PJ3WLRL2YRTQGQ6KRHU
	//
	plaintext := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	// Generate a SHA-256 hash of the plaintext token string. This will be the value
	// that we store in the `hash` field of our database table. Note that the
	// sha256.Sum256() function returns an *array* of length 32, so to make it easier to
	// work with we convert it to a slice using the [:] operator before storing it.
	hash := sha256.Sum256([]byte(plaintext))
	newHash := hash[:]

	return newHash, plaintext, nil
}

func (ah *AuthHandler) ForgotPasswordView(w http.ResponseWriter, r *http.Request) {
	data := ah.handler.NewTemplateData(r)
	data.PageTitle = "Forgot Password"

	auth.ForgotPasswordView(data).Render(r.Context(), w)
}

func (ah *AuthHandler) ForgotPasswordPostHanlder(w http.ResponseWriter, r *http.Request) {
	// handle form and its validation
	type ForgotPasswordForm struct {
		Email string `form:"email"`
	}

	var forgotPasswordForm ForgotPasswordForm

	err := ah.handler.DecodePostForm(r, &forgotPasswordForm)
	if err != nil {
		log.Panic(err)
	}

	// get the user by email
	user, err := ah.handler.DB.GetUserByEmail(r.Context(), forgotPasswordForm.Email)
	if err != nil {
		return
	}

	// handle the errors in the view

	// create token with ttl of 15min
	hash, plaintext, err := generateToken(int64(user.ID), 45*time.Second, ScopePasswordReset)
	if err != nil {
		println(err)
	}

	_, err = ah.handler.DB.CreateToken(r.Context(), queries.CreateTokenParams{
		UserID: int64(user.ID),
		Expiry: time.Now().Add(400),
		Scope:  ScopePasswordReset,
		Hash:   hash,
	})
	if err != nil {
		println(err)
	}

	// send the reset email with the token for the user
	data := map[string]any{
		"passwordResetLink": fmt.Sprintf("http://localhost:8080/reset-password?token=%s", plaintext),
	}
	err = ah.handler.Mailer.Send(user.Email, "reset_password.tmpl", data)
	if err != nil {
		ah.handler.Logger.PrintError(err, nil)
	}

	// set a flash message in the session manager

	ah.handler.SessionManager.Put(r.Context(), "flash", "Link sent to your email address")

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
