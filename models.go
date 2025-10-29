package main

import (
	"gorm.io/gorm"
)

type URLMapping struct {
	gorm.Model

	Code    string `gorm:"uniqueIndex;size:64;not null"`
	LongURL string `gorm:"not null"`

	Clicks uint `gorm:"default:0"`
}
