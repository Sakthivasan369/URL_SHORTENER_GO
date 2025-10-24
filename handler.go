package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// Global variable to hold the public domain (read from main.go/environment)
var BaseDomain = os.Getenv("SHORTENER_DOMAIN")

// --- Request/Response Structures ---

type ShortenRequest struct {
	LongURL string `json:"long_url"`
	Code    string `json:"code"` // optional custom alias
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"long_url"`
	Code     string `json:"code"`
}

// --- Handler Functions ---

// ShortenURLHandler handles POST /api/shorten to create a new short link.
func ShortenURLHandler(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// --- 1. Validate LongURL ---
	req.LongURL = strings.TrimSpace(req.LongURL)
	if req.LongURL == "" {
		http.Error(w, "Long URL cannot be empty.", http.StatusBadRequest)
		return
	}

	// --- 2. Determine Base Domain for Response ---
	domain := "http://" + r.Host
	if BaseDomain != "" {
		if !strings.HasPrefix(BaseDomain, "http://") && !strings.HasPrefix(BaseDomain, "https://") {
			domain = "https://" + BaseDomain
		} else {
			domain = BaseDomain
		}
	}

	code := strings.TrimSpace(req.Code)

	// --- 3. RESTRUCTURED LOGIC: Handle Custom Code vs. Generated Code ---

	if code == "" {
		// --- BRANCH A: NO CUSTOM CODE (Generate Random) ---

		// 1. Check if LongURL already exists (Idempotency Check)
		if existingMapping, err := FindURLMappingByLongURL(req.LongURL); err == nil {
			// Found existing mapping, return it
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(ShortenResponse{
				ShortURL: domain + "/" + existingMapping.Code,
				LongURL:  existingMapping.LongURL,
				Code:     existingMapping.Code,
			})
			return
		}

		// 2. Generate a new code
		code = GenerateShortCode()

	} else {
		// --- BRANCH B: CUSTOM CODE PROVIDED (Prioritize this) ---

		// 1. Validation Check (e.g., minimum length)
		if len(code) < 4 {
			http.Error(w, "Custom alias must be at least 4 characters long.", http.StatusBadRequest)
			return
		}

		// 2. Conflict Check (Check if custom code is already taken)
		if _, err := FindURLMappingByCode(code); err == nil {
			// Alias is already in use
			http.Error(w, "Custom alias '"+code+"' is already in use. Please choose another.", http.StatusConflict) // 409
			return
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			// Database error during lookup
			http.Error(w, "Database error during code check", http.StatusInternalServerError)
			return
		}
		// If we reach here, 'code' is the valid, available custom alias.
	}

	// --- 4. Save the new mapping ---
	mapping, err := SaveURLMapping(req.LongURL, code) // Calls the function in store.go
	if err != nil {
		http.Error(w, "Failed to save URL mapping", http.StatusInternalServerError)
		return
	}

	// --- 5. Respond with the created short URL ---
	w.WriteHeader(http.StatusCreated) // 201
	json.NewEncoder(w).Encode(ShortenResponse{
		ShortURL: domain + "/" + mapping.Code,
		LongURL:  mapping.LongURL,
		Code:     mapping.Code,
	})
}

// RedirectHandler (Unchanged)
func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Extract Code from the URL path
	code := chi.URLParam(r, "code")

	// 2. Lookup DB
	mapping, err := FindURLMappingByCode(code)

	// 3. Handle Not Found
	if errors.Is(err, gorm.ErrRecordNotFound) {
		http.Error(w, "Short URL Not Found", http.StatusNotFound) // 404
		return
	}
	if err != nil {
		http.Error(w, "Database Error", http.StatusInternalServerError)
		return
	}

	// 4. Increment Clicks
	IncrementClicks(mapping)

	// 5. Redirect to the LongURL (HTTP Status 302 Found)
	http.Redirect(w, r, mapping.LongURL, http.StatusFound)
}
