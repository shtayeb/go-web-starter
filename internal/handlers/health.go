package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

func (h *Handlers) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (h *Handlers) HealthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, _ := json.Marshal(h.DB.Health())

	h.Logger.PrintInfo("health route accessed", nil)

	_, _ = w.Write(jsonResp)
}
