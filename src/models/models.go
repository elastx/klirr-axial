package models

import (
	"time"
)

// User represents a user in the system, identified by their public key
type User struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time `json:"created_at"`
	Fingerprint string    `json:"fingerprint" gorm:"uniqueIndex"`
	PublicKey   string    `json:"public_key"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
}

// Message represents a message in the system
type Message struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time `json:"created_at"`
	Topic       string    `json:"topic"`
	Content     string    `json:"content"`
	Fingerprint string    `json:"fingerprint"`
} 