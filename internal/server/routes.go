package server

import (
	"net/http"

	"go-htmx-sqlite/cmd/web"
	"go-htmx-sqlite/internal/handlers"
	"go-htmx-sqlite/internal/handlers/auth"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// removes trailing slashed from the url
	r.Use(middleware.CleanPath)
	//  r.Use(s.secureHeaders)
	// r.Use(s.noSurf)
	r.Use(s.SessionManager.LoadAndSave)
	r.Use(s.authenticate)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// static file server
	fileServer := http.FileServer(http.FS(web.Files))
	r.Handle("/assets/*", fileServer)

	// s.Db is useless without the queries
	appHandlers := handlers.NewHandlers(s.Queries, s.Db, s.Logger, s.Mailer, s.SessionManager, s.Config)
	authHandlers := auth.NewAuthHandler(appHandlers)

	// No auth routes
	r.Group(func(r chi.Router) {
		// middleware
		r.Use(s.requireNoAuth)

		// Auth
		r.Get("/login", authHandlers.LoginViewHandler)
		r.Post("/login", authHandlers.LoginPostHandler)

		r.Get("/sign-up", authHandlers.SignUpViewHandler)
		r.Post("/sign-up", authHandlers.SignUpPostHandler)

		r.Get("/reset-password", authHandlers.ResetPasswordView)
		r.Post("/reset-password", authHandlers.ResetPasswordPostHandler)

		r.Get("/forgot-password", authHandlers.ForgotPasswordView)
		r.Post("/forgot-password", authHandlers.ForgotPasswordPostHanlder)
	})

	// Public routes
	r.Get("/", appHandlers.LandingViewHandler)
	r.Get("/health", appHandlers.HealthHandler)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(s.requireAuth)

		r.Post("/logout", authHandlers.LogoutPostHandler)

		r.Get("/dashboard", appHandlers.HelloFormHandler)
		r.Post("/hello", appHandlers.HelloWebHandler)
	})

	return r
}
