package server

import (
	"net/http"

	"go-htmx-sqlite/cmd/web"
	"go-htmx-sqlite/cmd/web/handlers"
	"go-htmx-sqlite/cmd/web/views"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (app *application) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/", handlers.HelloWorldHandler)

	r.Get("/health", handlers.HealthHandler)

	fileServer := http.FileServer(http.FS(web.Files))
	r.Handle("/assets/*", fileServer)

	r.Get("/web", templ.Handler(views.HelloForm()).ServeHTTP)
	r.Post("/hello", handlers.HelloWebHandler)

	return r
}
