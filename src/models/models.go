package models

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"gorm.io/gorm"
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
	ID          string    `json:"id" gorm:"primaryKey;column:message_id"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at"`
	Topic       string    `json:"topic,omitempty" gorm:"column:topic;default:null"`
	Recipient   string    `json:"recipient,omitempty" gorm:"column:recipient;default:null"`
	Content     string    `json:"content" gorm:"column:content;not null"`
	Author      string    `json:"fingerprint" gorm:"column:author;not null"` // Author's fingerprint
	Type        string    `json:"type" gorm:"column:type;not null;default:'bulletin'"` // 'private' or 'bulletin'
	ParentID    *string   `json:"parent_id,omitempty" gorm:"column:parent_message_id;default:null"` // Optional reference to parent message
}

// TableName specifies the table name for Message
func (Message) TableName() string {
	return "messages"
}

// GenerateMessageID creates a deterministic message ID based on the message properties
func (m *Message) GenerateMessageID() {
	// Create a string combining all relevant properties
	idString := fmt.Sprintf("%s|%s|%s|%s|%d",
		m.Author,
		m.Content,
		m.Topic+m.Recipient, // either topic (for bulletin) or recipient (for private)
		m.Type,
		m.CreatedAt.UnixNano(),
	)

	// Generate SHA-256 hash
	hash := sha256.Sum256([]byte(idString))
	
	// Convert to hex string
	m.ID = hex.EncodeToString(hash[:])
}

// BeforeCreate is called by GORM before creating a new message
func (m *Message) BeforeCreate(*gorm.DB) error {
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	m.GenerateMessageID()
	return nil
} 