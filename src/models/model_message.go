package models

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type CreateMessage struct {
	Content Crypto `json:"content" gorm:"column:content;not null"`
}

// Message represents a message in the system
type Message struct {
	Base
	Sender     Fingerprint   `json:"sender" gorm:"column:sender;type:text"`
	Recipients Fingerprints  `json:"recipients,omitempty" gorm:"column:recipients;type:jsonb"`
	CreateMessage
}

func (m *Message) In(messages []Message) bool {
	for _, message := range messages {
		if m.ID == message.ID {
			return true
		}
	}
	return false
}

// TableName specifies the table name for Message
func (Message) TableName() string {
	return "messages"
}

// Hash creates a deterministic message ID based on the message properties
func (m *Message) Hash() string {
	// Create a string combining all relevant properties
	recipients := ""
	for _, r := range m.Recipients {
		recipients += string(r)
	}
	hashStrings := []string{
		string(m.Sender),
		recipients,
		string(m.Content),
		m.CreatedAt.Format(time.RFC3339Nano),
	}

	idBytes := []byte{}
	for _, s := range hashStrings {
		idBytes = append(idBytes, []byte(s)...)
	}

	// Generate SHA-256 hash
	hash := sha256.Sum256([]byte(idBytes))

	// Convert to hex string
	return hex.EncodeToString(hash[:])
}

// BeforeCreate is called by GORM before creating a new message
func (m *Message) BeforeCreate(tx *gorm.DB) error {
	sender, recipients, encrypted, signed, err := m.Content.Analyze()
	if err != nil {
		return err
	}

	// Check for valid message properties
	if !signed {
		return fmt.Errorf("message must be signed")
	}
	if !encrypted {
		return fmt.Errorf("message must be encrypted")
	}

	// Check for tampered data during synchronization
	if m.Sender != "" && m.Sender != sender {
		return fmt.Errorf("message sender do not match content")
	} else {
		m.Sender = sender
	}
	if len(m.Recipients) != 0 {
		for _, r := range m.Recipients {
			found := false
			for _, rr := range recipients {
				if r == rr {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("message recipient %s not found in content", r)
			}
		}
		for _, rr := range recipients {
			found := false
			for _, r := range m.Recipients {
				if r == rr {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("message recipient %s found in content but not in recipients", rr)
			}
		}
	} else {
		m.Recipients = recipients
	}

	m.Base.ID = m.Hash()
	m.Base.BeforeCreate(tx)

	return nil
}
