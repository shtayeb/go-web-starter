package types

import "go-htmx-sqlite/internal/queries"

// TemplateData contains common data passed to all templates
type TemplateData struct {
	IsAuthenticated bool
	User            *queries.User
	CSRFToken       string
	Flash           string
	PageTitle       string
	Meta            map[string]string
}

// PageData wraps template data with page-specific data
type PageData struct {
	Template TemplateData
	Data     any
}
