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

	"github.com/markbates/goth"
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

// ProcessSocialAuth handles both login and signup for social authentication
func (as *AuthService) ProcessSocialAuth(ctx context.Context, gothUser goth.User, provider string) (*queries.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var user *queries.User
	var err error

	// Use transaction for consistency
	err = as.dbService.WithTransaction(ctx, func(tx *sql.Tx) error {
		qtx := as.dbQueries.WithTx(tx)

		// Check if user exists
		existingUser, userErr := qtx.GetUserByEmail(ctx, gothUser.Email)

		if userErr == nil {
			// User exists - verify provider linkage
			user, err = as.linkOrVerifyProvider(ctx, qtx, existingUser, gothUser, provider)
			if err != nil {
				return fmt.Errorf("failed to link provider: %w", err)
			}
		} else if errors.Is(userErr, sql.ErrNoRows) {
			// New user - create account
			user, err = as.createSocialUser(ctx, qtx, gothUser, provider)
			if err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
		} else {
			// Database error
			return fmt.Errorf("database error: %w", userErr)
		}

		// TODO: no need to store OAuth tokens in database
		// Update OAuth tokens
		if err := as.updateOAuthTokens(ctx, qtx, user.ID, gothUser, provider); err != nil {
			return fmt.Errorf("failed to update tokens: %w", err)
		}

		return nil
	})

	return user, err
}

// linkOrVerifyProvider handles existing users logging in with social auth
func (as *AuthService) linkOrVerifyProvider(
	ctx context.Context,
	qtx *queries.Queries,
	user queries.User,
	gothUser goth.User,
	provider string,
) (*queries.User, error) {
	// Check if this provider is already linked
	_, err := qtx.GetAccountByUserIdAndProvider(ctx, queries.GetAccountByUserIdAndProviderParams{
		UserID:     user.ID,
		ProviderID: sql.NullString{String: provider, Valid: true},
	})

	if err == nil {
		// Provider already linked - just return the user
		return &user, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// Provider not linked - for security, we should verify email ownership
	// In production, you might want to send a verification email
	// For now, we'll create the link if the email matches
	if user.Email != gothUser.Email {
		return nil, errors.New("email mismatch - cannot link provider")
	}

	// TODO: send verification email - user should verify their email

	// Create new account link
	_, err = qtx.CreateAccount(ctx, queries.CreateAccountParams{
		UserID:     user.ID,
		AccountID:  gothUser.UserID,
		ProviderID: sql.NullString{String: provider, Valid: true},
	})

	return &user, err
}

// createSocialUser creates a new user from social auth
func (as *AuthService) createSocialUser(
	ctx context.Context,
	qtx *queries.Queries,
	gothUser goth.User,
	provider string,
) (*queries.User, error) {
	// Create user
	user, err := qtx.CreateUser(ctx, queries.CreateUserParams{
		Name:      gothUser.Name,
		Email:     gothUser.Email,
		CreatedAt: time.Now().UTC(),
	})

	if err != nil {
		return nil, err
	}

	// Create account link
	_, err = qtx.CreateAccount(ctx, queries.CreateAccountParams{
		UserID:     user.ID,
		AccountID:  gothUser.UserID,
		ProviderID: sql.NullString{String: provider, Valid: true},
	})

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// updateOAuthTokens updates the OAuth tokens for a user
func (as *AuthService) updateOAuthTokens(
	ctx context.Context,
	qtx *queries.Queries,
	userID int32,
	gothUser goth.User,
	provider string,
) error {
	// Calculate token expiry time
	var expiresAt sql.NullTime
	if !gothUser.ExpiresAt.IsZero() {
		expiresAt = sql.NullTime{Time: gothUser.ExpiresAt, Valid: true}
	}

	// Update OAuth tokens
	err := qtx.UpdateAccountOAuthTokens(ctx, queries.UpdateAccountOAuthTokensParams{
		AccessToken:          sql.NullString{String: gothUser.AccessToken, Valid: gothUser.AccessToken != ""},
		RefreshToken:         sql.NullString{String: gothUser.RefreshToken, Valid: gothUser.RefreshToken != ""},
		AccessTokenExpiresAt: expiresAt,
		UserID:               userID,
		ProviderID:           sql.NullString{String: provider, Valid: true},
	})

	return err
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
		passwordValid = checkPasswordHash(account.Password.String, password)
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

func (as *AuthService) SignUp(ctx context.Context, name, email, password string, emailVerified bool) (*queries.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	user := &queries.User{}
	// DB transaction
	err := as.dbService.WithTransaction(ctx, func(tx *sql.Tx) error {
		qtx := as.dbQueries.WithTx(tx)

		// create user and handle DB errors - like user already exists
		user, err := qtx.CreateUser(ctx, queries.CreateUserParams{
			Name:          name,
			Email:         email,
			EmailVerified: emailVerified,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
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
			Password:  sql.NullString{String: hashedPassword, Valid: true},
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
		Password: sql.NullString{String: hashedPassword, Valid: true},
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

	if !checkPasswordHash(account.Password.String, currentPassword) {
		// invalid password - handle errors in login page
		return errors.New("invalid password")
	}

	hashedNewPassword, err := hashPassword(newPassword)
	if err != nil {
		return err
	}

	err = as.dbQueries.UpdateAccountPassword(ctx, queries.UpdateAccountPasswordParams{
		ID:       account.ID,
		Password: sql.NullString{String: hashedNewPassword, Valid: true},
	})

	return err
}
