// src/internal/api/server.go
// routes + handlers (POST /api/gec, /healthCheck) 
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"gec-demo/src/internal/gec"
	"gec-demo/src/internal/print"
)

// CORS middleware (simple + permissive for demo use).
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//TODO: If you later deploy this publicly, tighten this to your domain(s).
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// Endpoint: /healthCheck
func healthCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}

// Endpoint: POST /api/gec
func gecHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Allow "application/json" and "application/json; charset=utf-8"
	ct := r.Header.Get("Content-Type")
	if ct == "" || !strings.HasPrefix(strings.ToLower(ct), "application/json") {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	// Decode the request body
	var req gec.GecRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate the request
	if strings.TrimSpace(req.Text) == "" {
		http.Error(w, "Text field is required", http.StatusBadRequest)
		return
	}

	// Process the grammar check
	response, err := gec.MarkupGrammar(req.Text)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error processing grammar: %v", err), http.StatusInternalServerError)
		return
	}

	/* // Create response
	correctedText, markups := runGec(req.Text)
	response := GecResponse{
		CorrectedText: correctedText,
		TextMarkup:    markups,
	} */

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		print.Info("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func StartServer(port string) {
	if port == "" {
		port = "8089"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	
	// Routes
	http.HandleFunc("/api/gec", enableCORS(gecHandler))
	http.HandleFunc("/healthCheck", enableCORS(healthCheck))

	// Start the server
	print.Info("Server starting on port %s", port)
	print.Info("POST endpoint available at http://localhost%s/api/gec", port)
	print.Info("Health: http://localhost%s/healthCheck", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		print.Error("%v", err)
	}
}
