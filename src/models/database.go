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

const (
	UniqueViolationErr = "23505"
)

// InitDB connects to PostgreSQL, runs migrations, and returns the resulting
// *gorm.DB. The caller is responsible for holding the handle and passing it
// to the modules that need it — there is no package-level global.
func InitDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	log.Println("Running migrations...")
	if err := db.AutoMigrate(&User{}, &Message{}, &Bulletin{}); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %v", err)
	}

	var tableInfo []struct {
		ColumnName string `gorm:"column:column_name"`
		DataType   string `gorm:"column:data_type"`
		IsNullable string `gorm:"column:is_nullable"`
	}

	if err := db.Raw(`
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

	return db, nil
}

func IsDuplicateError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == UniqueViolationErr
}
