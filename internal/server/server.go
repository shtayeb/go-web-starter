package server

import (
	"fmt"
	"net/http"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"go-htmx-sqlite/internal/config"
	"go-htmx-sqlite/internal/database"
	"go-htmx-sqlite/internal/jsonlog"
	"go-htmx-sqlite/internal/mailer"
	"go-htmx-sqlite/internal/queries"
)

type Server struct {
	Port    int
	Db      database.Service
	Queries queries.Queries
	Mailer  mailer.Mailer
	Logger  *jsonlog.Logger
}

func NewServer() *http.Server {
	config := config.LoadConfigFromEnv()

	db := database.New(config.Database)

	s := &Server{
		Port:    config.Port,
		Db:      db,
		Queries: *queries.New(db.GetDB()),
		Logger:  jsonlog.New(os.Stdout, jsonlog.LevelInfo),
		Mailer:  mailer.New(config.Mailer),
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.Port),
		Handler:      s.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
