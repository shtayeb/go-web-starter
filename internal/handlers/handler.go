package handlers

import (
	"errors"
	"fmt"
	"go-htmx-sqlite/internal/config"
	"go-htmx-sqlite/internal/database"
	"go-htmx-sqlite/internal/jsonlog"
	"go-htmx-sqlite/internal/mailer"
	"go-htmx-sqlite/internal/queries"
	"go-htmx-sqlite/internal/types"
	"net/http"
	"runtime/debug"

	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/justinas/nosurf"
	"golang.org/x/crypto/bcrypt"
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

func (h *Handlers) getUser(r *http.Request) *queries.User {
	user, ok := r.Context().Value(config.UserContextKey).(queries.User)
	if !ok {
		return nil
	}

	return &user
}

func (h *Handlers) NewTemplateData(r *http.Request) types.TemplateData {
	return types.TemplateData{
		AppName:         h.Config.AppName,
		IsAuthenticated: h.isAuthenticated(r),
		User:            h.getUser(r),
		CSRFToken:       nosurf.Token(r),
		Flash:           h.SessionManager.PopString(r.Context(), "flash"),
	}
}

func (h *Handlers) newPageData(r *http.Request, data any) types.PageData {
	return types.PageData{
		Template: h.NewTemplateData(r),
		Data:     data,
	}
}

func (h *Handlers) DecodePostForm(r *http.Request, dst any) error {
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

// The serverError helper writes an error message and stack trace to the errorLog,
// then sends a generic 500 Internal Server Error response to the user.
func (h *Handlers) ServerError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	h.Logger.Write([]byte(trace))

	if h.Config.Debug {
		http.Error(w, trace, http.StatusInternalServerError)
		return
	}

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (h *Handlers) HashPassword(plainTextPassword string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), 14)
	return string(bytes), err
}
