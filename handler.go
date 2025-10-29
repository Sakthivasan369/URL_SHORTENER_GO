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

var BaseDomain = os.Getenv("SHORTENER_DOMAIN")

type ShortenRequest struct {
	LongURL string `json:"long_url"`
	Code    string `json:"code"`
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"long_url"`
	Code     string `json:"code"`
}

func ShortenURLHandler(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.LongURL = strings.TrimSpace(req.LongURL)
	if req.LongURL == "" {
		http.Error(w, "Long URL cannot be empty.", http.StatusBadRequest)
		return
	}

	domain := "http://" + r.Host
	if BaseDomain != "" {
		if !strings.HasPrefix(BaseDomain, "http://") && !strings.HasPrefix(BaseDomain, "https://") {
			domain = "https://" + BaseDomain
		} else {
			domain = BaseDomain
		}
	}

	code := strings.TrimSpace(req.Code)

	if code == "" {

		if existingMapping, err := FindURLMappingByLongURL(req.LongURL); err == nil {

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(ShortenResponse{
				ShortURL: domain + "/" + existingMapping.Code,
				LongURL:  existingMapping.LongURL,
				Code:     existingMapping.Code,
			})
			return
		}

		code = GenerateShortCode()

	} else {

		if len(code) < 4 {
			http.Error(w, "Custom alias must be at least 4 characters long.", http.StatusBadRequest)
			return
		}

		if _, err := FindURLMappingByCode(code); err == nil {

			http.Error(w, "Custom alias '"+code+"' is already in use. Please choose another.", http.StatusConflict) // 409
			return
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {

			http.Error(w, "Database error during code check", http.StatusInternalServerError)
			return
		}

	}

	mapping, err := SaveURLMapping(req.LongURL, code)
	if err != nil {
		http.Error(w, "Failed to save URL mapping", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ShortenResponse{
		ShortURL: domain + "/" + mapping.Code,
		LongURL:  mapping.LongURL,
		Code:     mapping.Code,
	})
}

func RedirectHandler(w http.ResponseWriter, r *http.Request) {

	code := chi.URLParam(r, "code")

	mapping, err := FindURLMappingByCode(code)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		http.Error(w, "Short URL Not Found", http.StatusNotFound) // 404
		return
	}
	if err != nil {
		http.Error(w, "Database Error", http.StatusInternalServerError)
		return
	}

	IncrementClicks(mapping)

	http.Redirect(w, r, mapping.LongURL, http.StatusFound)
}
