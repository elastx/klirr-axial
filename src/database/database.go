package database

import (
	"fmt"

	"axial/config"
	"axial/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Connect establishes a connection to the database and performs migrations
func Connect(cfg config.DatabaseConfig) error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// Run migrations
	err = DB.AutoMigrate(&models.User{}, &models.Message{})
	if err != nil {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	return nil
} 