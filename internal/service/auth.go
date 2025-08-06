package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"errors"
	"fmt"
	"go-htmx-sqlite/internal/config"
	"go-htmx-sqlite/internal/database"
	"go-htmx-sqlite/internal/queries"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	dbQueries *queries.Queries
	dbService database.Service
}

func NewAuthService(dbQueries *queries.Queries, db database.Service) *AuthService {
	return &AuthService{
		dbQueries: dbQueries,
		dbService: db,
	}
}

func checkPasswordHash(hashedPassword, plainTextPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainTextPassword))
	return err == nil
}

// TODO: duplicate func, move it to util package
func hashPassword(plainTextPassword string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), 14)
	return string(bytes), err
}

func (as *AuthService) GenerateToken(ctx context.Context, userID int64, ttl time.Duration, scope string) (string, error) {
	// Initialize a zero-valued byte slice with a length of 16 bytes.
	randomBytes := make([]byte, 16)

	// Use the Read() function from the crypto/rand package to fill the byte slice with
	// random bytes from your operating system's CSPRNG.
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
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

	_, err = as.dbQueries.CreateToken(ctx, queries.CreateTokenParams{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
		Hash:   newHash,
	})

	if err != nil {
		return "", err
	}

	return plaintext, nil
}

func (as *AuthService) Login(ctx context.Context, email string, password string) (*queries.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	var user queries.User
	var account queries.Account
	var userErr, accountErr error

	// Always perform both operations to prevent timing attacks
	user, userErr = as.dbQueries.GetUserByEmail(ctx, email)
	if userErr == nil {
		account, accountErr = as.dbQueries.GetAccountByUserId(ctx, user.ID)
	}

	// Always perform password check, even with dummy hash to prevent timing attacks
	var passwordValid bool
	if userErr == nil && accountErr == nil {
		passwordValid = checkPasswordHash(account.Password, password)
	} else {
		// Perform dummy hash check to maintain constant time
		checkPasswordHash("$2a$14$dummy.hash.to.prevent.timing.attacks.abcdefghijklmnopqrstuvwxyz", password)
		passwordValid = false
	}

	if userErr != nil || accountErr != nil || !passwordValid {
		return nil, errors.New("invalid email or password")
	}

	return &user, nil
}

func (as *AuthService) SignUp(ctx context.Context, name, email, password string) (*queries.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	user := &queries.User{}
	// DB transaction
	err := as.dbService.WithTransaction(ctx, func(tx *sql.Tx) error {
		qtx := as.dbQueries.WithTx(tx)

		// create user and handle DB errors - like user already exists
		user, err := qtx.CreateUser(ctx, queries.CreateUserParams{
			Name:      name,
			Email:     email,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		})
		if err != nil {
			return err
		}

		// hash the password
		hashedPassword, err := hashPassword(password)
		if err != nil {
			// handle error in the view
			return err
		}

		// create account - and handle errors
		_, err = qtx.CreateAccount(ctx, queries.CreateAccountParams{
			UserID:    user.ID,
			AccountID: user.Name,
			Password:  hashedPassword,
		})

		return err
	})

	return user, err
}

func (as *AuthService) GetPasswordResetLink(ctx context.Context, email string, baseURL string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// check if user exists by their email
	user, err := as.dbQueries.GetUserByEmail(ctx, email)
	if err != nil {
		// handle error
		log.Println(err)
		return "", err
	}

	// create token with ttl of 45min
	plaintext, err := as.GenerateToken(ctx, int64(user.ID), 45*time.Minute, config.ScopePasswordReset)
	if err != nil {
		println(err)
		return "", err
	}

	passwordResetLink := fmt.Sprintf("%s/reset-password?token=%s", baseURL, plaintext)

	return passwordResetLink, nil
}

func (as *AuthService) GetValidTokenUser(ctx context.Context, token string) (*queries.User, error) {
	// hash the plainText token
	tokenHash := sha256.Sum256([]byte(token))

	// compare the token with the hashed one in the database
	user, err := as.dbQueries.GetUserByToken(ctx, queries.GetUserByTokenParams{
		Hash:   tokenHash[:],
		Scope:  config.ScopePasswordReset,
		Expiry: time.Now(),
	})

	return &user.User, err
}

func (as *AuthService) ResetPassword(ctx context.Context, token, password string) (*queries.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// get user
	user, err := as.GetValidTokenUser(ctx, token)
	if err != nil {
		return nil, err
	}

	// get user account
	account, err := as.dbQueries.GetAccountByUserId(ctx, user.ID)
	if err != nil {
		return user, err
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return user, err
	}

	err = as.dbQueries.UpdateAccountPassword(ctx, queries.UpdateAccountPasswordParams{
		ID:       account.ID,
		Password: hashedPassword,
	})
	if err != nil {
		return user, err
	}

	// delete token
	err = as.dbQueries.DeleteAllForUser(ctx, queries.DeleteAllForUserParams{
		Scope:  config.ScopePasswordReset,
		UserID: int64(user.ID),
	})

	return user, err
}

func (as *AuthService) UpdateUserNameAndImage(ctx context.Context, id int32, name, image string) (queries.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	user, err := as.dbQueries.UpdateUserNameAndImage(ctx, queries.UpdateUserNameAndImageParams{
		ID:    id,
		Name:  name,
		Image: sql.NullString{String: image, Valid: image != ""},
	})

	return user, err
}

func (as *AuthService) UpdateAccountPassword(ctx context.Context, userId int32, currentPassword, newPassword string) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// get user account
	account, err := as.dbQueries.GetAccountByUserId(ctx, userId)
	if err != nil {
		return err
	}

	if !checkPasswordHash(account.Password, currentPassword) {
		// invalid password - handle errors in login page
		return errors.New("invalid password")
	}

	hashedNewPassword, err := hashPassword(newPassword)
	if err != nil {
		return err
	}

	err = as.dbQueries.UpdateAccountPassword(ctx, queries.UpdateAccountPasswordParams{
		ID:       account.ID,
		Password: hashedNewPassword,
	})

	return err
}
