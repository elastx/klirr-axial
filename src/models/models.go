package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system, identified by their public key
type User struct {
	gorm.Model
	Fingerprint string `gorm:"uniqueIndex;not null"` // Public key fingerprint
	PublicKey   string `gorm:"not null"`            // Full public key
}

// Message represents a message in the system
type Message struct {
	gorm.Model
	MessageID  string    `gorm:"uniqueIndex;not null"` // Hash of author, recipient, and content
	Author     string    `gorm:"index;not null"`       // Author's public key fingerprint
	Recipient  string    `gorm:"index;not null"`       // Either public key fingerprint or thread topic
	Content    string    `gorm:"not null"`             // Message content (may be encrypted)
	CreatedAt  time.Time `gorm:"index;not null"`       // When the message was created on the source device
	HeaderType string    `gorm:"not null"`             // Encryption method/version used (if encrypted)
} 