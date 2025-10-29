package main

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set. Cannot connect to Supabase.")
	}

	var err error

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to postgres: %v", err)
	}

	if err := DB.AutoMigrate(&URLMapping{}); err != nil {
		log.Fatalf("GORM auto migrate failed: %v", err)
	}
	log.Println("Successfully connected to Supabase Postgres and migrated database.")
}

func FindURLMappingByCode(code string) (*URLMapping, error) {
	var mapping URLMapping
	result := DB.Where("code = ?", code).First(&mapping)
	if result.Error != nil {

		return nil, result.Error
	}
	return &mapping, nil
}

func FindURLMappingByLongURL(longURL string) (*URLMapping, error) {
	var mapping URLMapping
	result := DB.Where("long_url = ?", longURL).First(&mapping)
	if result.Error != nil {
		return nil, result.Error
	}
	return &mapping, nil
}

func SaveURLMapping(longURL, code string) (*URLMapping, error) {
	mapping := URLMapping{
		Code:    code,
		LongURL: longURL,
	}

	if err := DB.Create(&mapping).Error; err != nil {
		return nil, err
	}
	return &mapping, nil
}

func IncrementClicks(mapping *URLMapping) {

	DB.Model(mapping).Update("Clicks", gorm.Expr("Clicks + ?", 1))
}

func GenerateShortCode() string {

	return "s" + time.Now().Format("20060102150405")
}
