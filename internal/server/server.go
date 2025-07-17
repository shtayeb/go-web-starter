package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"go-htmx-sqlite/internal/database"
	"go-htmx-sqlite/internal/mailer"
)

type Server struct {
	port   int
	db     database.Service
	mailer mailer.Mailer
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

	NewServer := &Server{
		port:   port,
		db:     database.New(),
		mailer: mailer.New(smtp.host, smtp.port, smtp.username, smtp.password, smtp.sender),
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
