package server

import (
	"net/http"
	"time"

	"go-web-starter/cmd/web"
	"go-web-starter/internal/handlers"
	"go-web-starter/internal/handlers/auth"
	"go-web-starter/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// removes trailing slashed from the url
	r.Use(middleware.CleanPath)
	//  r.Use(s.secureHeaders)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(s.noSurf)
	r.Use(s.SessionManager.LoadAndSave)
	r.Use(s.authenticate)

	// static file server
	fileServer := http.FileServer(http.FS(web.Files))
	r.Handle("/assets/*", fileServer)

	// s.Db is useless without the queries
	appHandlers := handlers.NewHandlers(s.Queries, s.Db, s.Logger, s.Mailer, s.SessionManager, s.Config)

	authService := service.NewAuthService(&s.Queries, s.Db)
	authHandlers := auth.NewAuthHandler(appHandlers, authService)

	// No auth routes
	r.With(
		//middlewares
		s.requireNoAuth,
		httprate.LimitByIP(100, 1*time.Minute),
	).Group(func(r chi.Router) {
		// Auth
		r.Get("/login", authHandlers.LoginViewHandler)
		r.Post("/login", authHandlers.LoginPostHandler)

		r.Get("/signup", authHandlers.SignUpViewHandler)
		r.Post("/signup", authHandlers.SignUpPostHandler)

		r.Get("/reset-password", authHandlers.ResetPasswordView)
		r.Post("/reset-password", authHandlers.ResetPasswordPostHandler)

		r.Get("/forgot-password", authHandlers.ForgotPasswordView)
		r.Post("/forgot-password", authHandlers.ForgotPasswordPostHanlder)

		// social logins
		r.Get("/auth/{provider}", authHandlers.SocialAuthHandler)
		r.Get("/auth/{provider}/callback", authHandlers.SocialAuthCallbackHandler)
	})

	// Public routes
	r.Get("/", appHandlers.LandingViewHandler)
	r.Get("/health", appHandlers.HealthHandler)

	// Protected routes
	r.With(
		//middlewares
		s.requireAuth,
	).Group(func(r chi.Router) {
		r.Post("/logout", authHandlers.LogoutPostHandler)

		r.Get("/profile", authHandlers.ProfileViewHandler)
		r.Post("/profile/update", authHandlers.UpdateUserNameAndImageHandler)
		r.Post("/profile/update-password", authHandlers.UpdateAccountPasswordHandler)
		r.Post("/profile/delete-account", authHandlers.DeleteAccountHandler)

		r.Get("/projects", appHandlers.ProjectViewHandler)

		r.Get("/dashboard", appHandlers.DashboardViewHandler)
		r.Post("/hello", appHandlers.HelloWebHandler)
	})

	return r
}
