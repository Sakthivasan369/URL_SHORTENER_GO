package main

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB connects to PostgreSQL and runs GORM AutoMigrate.
func InitDB() {
	// 1. Get DSN from environment variable (Supabase connection string)
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set. Cannot connect to Supabase.")
	}

	var err error
	// 2. Open the connection using the postgres driver
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to postgres: %v", err)
	}

	// 3. AutoMigrate (creates table if not exists)
	if err := DB.AutoMigrate(&URLMapping{}); err != nil {
		log.Fatalf("GORM auto migrate failed: %v", err)
	}
	log.Println("Successfully connected to Supabase Postgres and migrated database.")
}

// FindURLMappingByCode looks up the LongURL for a given short code.
func FindURLMappingByCode(code string) (*URLMapping, error) {
	var mapping URLMapping
	result := DB.Where("code = ?", code).First(&mapping)
	if result.Error != nil {
		// gorm.ErrRecordNotFound is returned if no record matches
		return nil, result.Error
	}
	return &mapping, nil
}

// FindURLMappingByLongURL checks if a LongURL already exists.
func FindURLMappingByLongURL(longURL string) (*URLMapping, error) {
	var mapping URLMapping
	result := DB.Where("long_url = ?", longURL).First(&mapping)
	if result.Error != nil {
		return nil, result.Error
	}
	return &mapping, nil
}

// SaveURLMapping inserts a new URLMapping record into the database.
func SaveURLMapping(longURL, code string) (*URLMapping, error) {
	mapping := URLMapping{
		Code:    code,
		LongURL: longURL,
		// Clicks is defaulted to 0 by the model tag
	}
	// Check for unique index violation (Code collision)
	if err := DB.Create(&mapping).Error; err != nil {
		return nil, err
	}
	return &mapping, nil
}

// IncrementClicks updates the click counter atomically.
func IncrementClicks(mapping *URLMapping) {
	// Uses GORM's expression to safely increment the counter
	DB.Model(mapping).Update("Clicks", gorm.Expr("Clicks + ?", 1))
}

// --- Helper for generating a unique short code ---
// NOTE: For a real app, this should be more robust (e.g., using a library
// or ensuring true cryptographic randomness).
func GenerateShortCode() string {
	// Simple time-based code generation for example
	return "s" + time.Now().Format("20060102150405")
}
