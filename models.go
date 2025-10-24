package main

import (
	"gorm.io/gorm"
)

// URLMapping stores the mapping between short code and long URL
type URLMapping struct {
	gorm.Model
	// Code will be indexed and unique for fast lookups.
	Code    string `gorm:"uniqueIndex;size:64;not null"`
	LongURL string `gorm:"not null"`
	// Clicks tracks usage.
	Clicks uint `gorm:"default:0"`
}

// NOTE: When connecting to Supabase, GORM will create a table named 'url_mappings'
// based on the struct name 'URLMapping'.
