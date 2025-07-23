package handlers

import (
	"go-htmx-sqlite/internal/database"
	"go-htmx-sqlite/internal/jsonlog"
	"go-htmx-sqlite/internal/mailer"
	"go-htmx-sqlite/internal/queries"
)

type Handlers struct {
	DB        queries.Queries
	DbService database.Service
	Logger    *jsonlog.Logger
	Mailer    mailer.Mailer
}

func NewHandlers(q queries.Queries, dbService database.Service, logger *jsonlog.Logger, mailer mailer.Mailer) *Handlers {
	return &Handlers{
		DB:        q,
		DbService: dbService,
		Logger:    logger,
		Mailer:    mailer,
	}
}
