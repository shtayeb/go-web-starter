package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"go-htmx-sqlite/internal/database"
	"go-htmx-sqlite/internal/jsonlog"
	"go-htmx-sqlite/internal/mailer"
)

type Server struct {
	Port   int
	Db     database.Service
	Mailer mailer.Mailer
	Logger *jsonlog.Logger
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	smtp := struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}{
		host:     os.Getenv("SMTP_HOST"),
		port:     func() int { p, _ := strconv.Atoi(os.Getenv("SMTP_PORT")); return p }(),
		username: os.Getenv("SMTP_USERNAME"),
		password: os.Getenv("SMTP_PASSWORD"),
		sender:   os.Getenv("SMTP_SENDER"),
	}

	s := &Server{
		Port:   port,
		Db:     database.New(),
		Logger: jsonlog.New(os.Stdout, jsonlog.LevelInfo),
		Mailer: mailer.New(smtp.host, smtp.port, smtp.username, smtp.password, smtp.sender),
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
