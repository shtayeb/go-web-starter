package handlers

import (
	"errors"
	"fmt"
	"go-htmx-sqlite/internal/config"
	"go-htmx-sqlite/internal/database"
	"go-htmx-sqlite/internal/jsonlog"
	"go-htmx-sqlite/internal/mailer"
	"go-htmx-sqlite/internal/queries"
	"net/http"
	"runtime/debug"

	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/justinas/nosurf"
)

type Handlers struct {
	DB             queries.Queries
	DbService      database.Service
	Logger         *jsonlog.Logger
	Mailer         mailer.Mailer
	SessionManager *scs.SessionManager
	Config         config.Config
}

func NewHandlers(
	q queries.Queries,
	dbService database.Service,
	logger *jsonlog.Logger,
	mailer mailer.Mailer,
	sessionManager *scs.SessionManager,
	config config.Config,
) *Handlers {
	return &Handlers{
		DB:             q,
		DbService:      dbService,
		Logger:         logger,
		Mailer:         mailer,
		SessionManager: sessionManager,
		Config:         config,
	}
}

// helpers

func (h *Handlers) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(config.IsAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}

	return isAuthenticated
}

func (h *Handlers) decodePostForm(r *http.Request, dst any) error {
	// Call ParseForm() on the request
	err := r.ParseForm()
	if err != nil {
		return err
	}

	decoder := form.NewDecoder()

	err = decoder.Decode(dst, r.PostForm)
	if err != nil {
		// If we try to use an invalid target destination, the Decode() method
		// will return an error with the type *form.InvalidDecoderError.We use
		// errors.As() to check for this and raise a panic rather than returning the error
		var invalideDecoderError *form.InvalidDecoderError

		if errors.As(err, &invalideDecoderError) {
			panic(err)
		}

		// return as normal
		return err
	}

	return nil
}

type TemplateData struct {
	// CurrentYear     int
	// Snippet         *models.Snippet
	// Snippets        []*models.Snippet
	// Form            any
	Flash           string
	IsAuthenticated bool
	CSRFToken       string
	User            queries.User
}

func (h *Handlers) newTemplateData(r *http.Request) *TemplateData {
	return &TemplateData{
		Flash:           h.SessionManager.PopString(r.Context(), "flash"),
		IsAuthenticated: h.isAuthenticated(r),
		CSRFToken:       nosurf.Token(r),
		User:            r.Context().Value("user").(queries.User),
	}

}

// The serverError helper writes an error message and stack trace to the errorLog,
// then sends a generic 500 Internal Server Error response to the user.
func (h *Handlers) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	h.Logger.Write([]byte(trace))

	if h.Config.Debug {
		http.Error(w, trace, http.StatusInternalServerError)
		return
	}

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
