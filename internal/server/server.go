package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"

	"go-web-starter/internal/config"
	"go-web-starter/internal/database"
	"go-web-starter/internal/jsonlog"
	"go-web-starter/internal/mailer"
	"go-web-starter/internal/queries"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/sqlite3store"
)

type Server struct {
	Port           int
	Db             database.Service
	Queries        queries.Queries
	Mailer         mailer.Mailer
	Logger         *jsonlog.Logger
	SessionManager *scs.SessionManager
	Config         config.Config
}

func NewServer(cfg config.Config, db database.Service, q *queries.Queries, logger *jsonlog.Logger, mailer mailer.Mailer, sessionManager *scs.SessionManager) *Server {
	s := &Server{
		Port:           cfg.Port,
		Db:             db,
		Queries:        *q,
		Logger:         logger,
		Mailer:         mailer,
		SessionManager: sessionManager,
		Config:         cfg,
	}

	return s
}

func NewSessionManager(db *sql.DB, dbType string) *scs.SessionManager {
	sessionManager := scs.New()

	switch dbType {
	case "sqlite", "sqlite3":
		sessionManager.Store = sqlite3store.New(db)
	case "postgres", "postgresql":
		sessionManager.Store = postgresstore.New(db)
	}

	sessionManager.Lifetime = 12 * time.Hour
	// Make sure that the Secure attribute is set on our session cookies. Setting this means that the cookie will only be sent by a user's web
	// browser when a HTTPS connection is being used (and won't be sent over an unsecure HTTP connection).
	sessionManager.Cookie.Secure = true

	return sessionManager
}

func NewHttpServer() *http.Server {
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// load the .env file. by default, it will load the .env file in the root directory
	err := godotenv.Load()
	if err != nil {
		logger.Info(fmt.Sprintf("Error loading .env file - using default values now: %v", err))
	}

	config := config.LoadConfigFromEnv()

	dbService := database.New(config.Database)
	sqlDb := dbService.GetDB()

	goth.UseProviders(
		google.New(
			config.SocialLogins.GoogleClientID,
			config.SocialLogins.GoogleClientSecret,
			fmt.Sprintf("%s/auth/google/callback", config.AppURL),
		),
	)

	s := NewServer(
		config,
		dbService,
		queries.New(sqlDb),
		logger,
		mailer.New(config.Mailer),
		NewSessionManager(sqlDb, config.Database.Type),
	)

	// Declare Server config
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.Port),
		Handler:      s.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return httpServer
}
