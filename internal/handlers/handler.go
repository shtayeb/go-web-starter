package handlers

import (
	"go-htmx-sqlite/internal/database"
	"go-htmx-sqlite/internal/jsonlog"
	"go-htmx-sqlite/internal/mailer"
)

type Handlers struct {
	DB     database.Service
	Logger *jsonlog.Logger
	Mailer mailer.Mailer
}

func NewHandlers(db database.Service, logger *jsonlog.Logger, mailer mailer.Mailer) *Handlers {
	return &Handlers{
		DB:     db,
		Logger: logger,
		Mailer: mailer,
	}
}
