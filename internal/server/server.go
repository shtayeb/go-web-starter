package server

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	_ "github.com/joho/godotenv/autoload"

	"go-htmx-sqlite/internal/config"
	"go-htmx-sqlite/internal/database"
	"go-htmx-sqlite/internal/jsonlog"
	"go-htmx-sqlite/internal/mailer"
	"go-htmx-sqlite/internal/queries"

	"github.com/alexedwards/scs/postgresstore"
)

type Server struct {
	Port           int
	Db             database.Service
	Queries        queries.Queries
	Mailer         mailer.Mailer
	Logger         *jsonlog.Logger
	SessionManager *scs.SessionManager
}

func NewServer() *http.Server {
	config := config.LoadConfigFromEnv()

	db := database.New(config.Database)

	sessionManager := scs.New()
	sessionManager.Store = postgresstore.New(db.GetDB())
	sessionManager.Lifetime = 12 * time.Hour
	// Make sure that the Secure attribute is set on our session cookies. Setting this means that the cookie will only be sent by a user's web
	// browser when a HTTPS connection is being used (and won't be sent over an unsecure HTTP connection).
	sessionManager.Cookie.Secure = true

	s := &Server{
		Port:           config.Port,
		Db:             db,
		Queries:        *queries.New(db.GetDB()),
		Logger:         jsonlog.New(os.Stdout, jsonlog.LevelInfo),
		Mailer:         mailer.New(config.Mailer),
		SessionManager: sessionManager,
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
