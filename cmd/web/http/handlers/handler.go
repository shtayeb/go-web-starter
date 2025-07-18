package handlers

import "go-htmx-sqlite/internal/server"

type Handlers struct {
	*server.Server
}

func NewHandlers(app *server.Server) *Handlers {
	return &Handlers{app}
}
