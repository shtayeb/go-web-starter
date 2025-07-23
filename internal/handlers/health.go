package handlers

import (
	"encoding/json"
	"net/http"
)

func (h *Handlers) HealthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, _ := json.Marshal(h.DbService.Health())

	h.Logger.PrintInfo("health route accessed", nil)

	_, _ = w.Write(jsonResp)
}
