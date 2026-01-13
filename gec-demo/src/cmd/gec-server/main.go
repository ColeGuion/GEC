// src/cmd/gec-server/main.go
// Starts HTTP server on :8089, serves /api/gec + static files 
package gecserver

import (
	"os"

	"gec-demo/src/internal/api"
)


// Entry point for the GEC server binary.
// Reads PORT from env (defaults to 8089) and starts the HTTP server.
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8089"
	}

	api.StartServer(port)
}