package handlers

import (
	"go-web-starter/cmd/web/views"
	"log"
	"net/http"
)

func (h *Handlers) LandingViewHandler(w http.ResponseWriter, r *http.Request) {
	data := h.NewTemplateData(r)
	data.PageTitle = "Welcome"

	views.LandingView(data).Render(r.Context(), w)
}

func (h *Handlers) DashboardViewHandler(w http.ResponseWriter, r *http.Request) {
	data := h.NewTemplateData(r)
	data.PageTitle = "Dashboard"

	views.DashboardView(data).Render(r.Context(), w)
}

func (h *Handlers) ProjectViewHandler(w http.ResponseWriter, r *http.Request) {
	data := h.NewTemplateData(r)
	data.PageTitle = "Projects"

	views.ProjectView(data).Render(r.Context(), w)
}

func (h *Handlers) HelloWebHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}

	name := r.FormValue("name")
	component := views.HelloPost(name)

	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Fatalf("Error rendering in HelloWebHandler: %e", err)
	}
}
