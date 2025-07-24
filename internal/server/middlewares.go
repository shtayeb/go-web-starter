package server

import (
	"context"
	"net/http"

	"github.com/justinas/nosurf"
)

type contextKey string

const (
	isAuthenticatedContextKey = contextKey("isAuthenticated")
	userContextKey            = contextKey("user")
)

func (s *Server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// sessionManager
		id := s.SessionManager.GetInt32(r.Context(), "authenticatedUserID")
		if id == 0 {
			http.Redirect(w, r.WithContext(r.Context()), "/login", http.StatusSeeOther)
			return
		}

		user, err := s.Queries.GetUserById(r.Context(), id)
		if err != nil {
			http.Redirect(w, r.WithContext(r.Context()), "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), isAuthenticatedContextKey, true)
		ctx = context.WithValue(ctx, userContextKey, user)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (s *Server) noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})

	return csrfHandler
}

func (s *Server) secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")

		// Any code above ServeHttp will be executed on the way down the chain
		next.ServeHTTP(w, r)
		// Any code here will execute on the way back up the chain.
	})
}
