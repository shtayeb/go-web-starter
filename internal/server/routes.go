package server

import (
	"net/http"

	"go-htmx-sqlite/cmd/web"
	"go-htmx-sqlite/cmd/web/views"
	"go-htmx-sqlite/internal/handlers"

	"github.com/a-h/templ"
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

	// s.Db is useless without the queries
	appHandlers := handlers.NewHandlers(s.Queries, s.Db, s.Logger, s.Mailer, s.SessionManager, s.Config)

	// static file server
	fileServer := http.FileServer(http.FS(web.Files))
	r.Handle("/assets/*", fileServer)

	// No auth routes
	r.Group(func(r chi.Router) {
		// middleware
		r.Use(s.requireNoAuth)

		// Auth
		r.Get("/login", appHandlers.LoginViewHandler)
		r.Post("/login", appHandlers.LoginPostHandler)

		r.Get("/sign-up", appHandlers.SignUpViewHandler)
		r.Post("/sign-up", appHandlers.SignUpPostHandler)

		r.Get("/reset-password", appHandlers.ResetPasswordView)
		r.Get("/forgot-password", appHandlers.ForgotPasswordView)
	})

	// Public routes
	r.Get("/", appHandlers.LandingViewHandler)
	r.Get("/health", appHandlers.HealthHandler)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(s.requireAuth)

		r.Post("/logout", appHandlers.LogoutPostHandler)

		r.Get("/dashboard", templ.Handler(views.HelloForm()).ServeHTTP)
		r.Post("/hello", appHandlers.HelloWebHandler)
	})

	return r
}
