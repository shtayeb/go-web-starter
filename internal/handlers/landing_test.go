package handlers

import (
	"go-web-starter/internal/tests"
	"log"
	"testing"
)

func TestLandingViewHandler(t *testing.T) {
	server := tests.StartTestServer(t)

	log.Println("Running TestLandingViewHandler", server.Config)
}
