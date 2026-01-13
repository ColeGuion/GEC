package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// CORS middleware
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

func runGec(text string) (string, []Markup) {
	// For demonstration, we'll create some dummy edits
	edits := []Markup{
		{
			Index:   0,
			Length:  2,
			Message: "Change the capitalization “We”",
			Type:    "GRAMMAR_SUGGESTION",
		},
		{
			Index:   3,
			Length:  5,
			Message: "Possible spelling mistake found.",
			Type:    "Spelling Mistake",
		},
		{
			Index:   13,
			Length:  2,
			Message: "Did you mean “a”?",
			Type:    "GRAMMAR_SUGGESTION",
		},
	}

	// Simple correction (in reality, you'd apply the edits properly)
	//corrected := text + " [grammar corrected]"
	corrected := "We should buy a car."

	return corrected, edits
}

// GEC Endpoint Handler
func gecHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check content type
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	// Decode the request body
	var req GecRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate the request
	if req.Text == "" {
		http.Error(w, "Text field is required", http.StatusBadRequest)
		return
	}

	// Process the grammar check
	response, err := MarkupGrammar(req.Text)
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
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func StartServer() {
	// Register the handler with CORS middleware
	http.HandleFunc("/api/gec", enableCORS(gecHandler))

	// Start the server
	port := ":8089"
	log.Printf("Server starting on port %s", port)
	log.Printf("POST endpoint available at http://localhost%s/api/gec", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
