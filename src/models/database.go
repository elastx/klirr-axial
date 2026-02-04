package models

import (
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"axial/config"
)

var DB *gorm.DB

const (
	UniqueViolationErr = "23505"
)

// InitDB establishes a connection to the database and performs migrations
func InitDB(cfg config.DatabaseConfig) error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)

	// Enable detailed logging for migrations
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	log.Println("Running migrations...")
	// Run migrations
	if err := DB.AutoMigrate(&User{}, &Group{}, &Message{}, &Bulletin{}, &File{}); err != nil {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	// Debug: Print table schema
	var tableInfo []struct {
		ColumnName string `gorm:"column:column_name"`
		DataType   string `gorm:"column:data_type"`
		IsNullable string `gorm:"column:is_nullable"`
	}
	
	if err := DB.Raw(`
		SELECT column_name, data_type, is_nullable 
		FROM information_schema.columns 
		WHERE table_name = 'bulletin_board'
		ORDER BY ordinal_position
	`).Scan(&tableInfo).Error; err != nil {
		log.Printf("Failed to get table info: %v", err)
	} else {
		log.Println("Bulletin table schema:")
		for _, col := range tableInfo {
			log.Printf("  %s (%s, nullable: %s)", col.ColumnName, col.DataType, col.IsNullable)
		}
	}

	return nil
} 

func GetUserByFingerprint(fingerprint Fingerprint) (*User, error) {
	var user User
	if err := DB.Where("fingerprint = ?", fingerprint).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func IsDuplicateError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == UniqueViolationErr
}