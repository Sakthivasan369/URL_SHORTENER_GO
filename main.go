package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors" // <-- THIS LINE WAS ADDED
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  No .env file found or error loading it (if in production, ignore this).")
	}
	InitDB()

	r := chi.NewRouter()

	// Use common middlewares (request logging, panic recovery)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// --- ADD CORS CONFIGURATION HERE ---
	r.Use(cors.Handler(cors.Options{
		// Allow requests from your React development server
		AllowedOrigins: []string{"http://localhost:5173"},
		// Allow POST and GET methods (and OPTIONS for pre-flight)
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		// Allow standard headers
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		// Allow credentials (if you were using cookies/sessions)
		AllowCredentials: true,
		// MaxAge: cache preflight requests for 300 seconds
		MaxAge: 300,
	}))
	// --- END CORS CONFIGURATION ---

	// --- Define Routes ---
	r.Post("/api/shorten", ShortenURLHandler)
	r.Get("/{code}", RedirectHandler)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
