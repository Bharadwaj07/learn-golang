package main

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Link struct {
	ID        uint   `gorm:"primaryKey"`
	Code      string `gorm:"uniqueIndex;not null"`
	URL       string `gorm:"not null"`
	CreatedAt time.Time
}

func openDB() (*gorm.DB, error) {
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DB_URL environment variable is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	if err := db.AutoMigrate(&Link{}); err != nil {
		return nil, fmt.Errorf("migrating database: %w", err)
	}

	return db, nil
}

func saveLink(db *gorm.DB, code, url string) error {
	link := Link{Code: code, URL: url}
	result := db.Create(&link)
	return result.Error
}

func getLink(db *gorm.DB, code string) (string, error) {
	var link Link
	result := db.Where("code = ?", code).First(&link)
	if result.Error != nil {
		return "", result.Error
	}
	return link.URL, nil
}
