package config

type contextKey string

const (
	IsAuthenticatedContextKey = contextKey("isAuthenticated")
	AuthenticatedUserID       = contextKey("authenticatedUserID")
	UserContextKey            = contextKey("user")
)
