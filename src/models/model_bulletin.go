package models

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type CreateBulletin struct {
	Topic string `json:"topic" gorm:"column:topic;not null"`
	Content Crypto `json:"content" gorm:"column:content;not null"`
	ParentID *string `json:"parent_id,omitempty" gorm:"column:parent_id;default:null"`
}

type Bulletin struct {
	Base
	Sender Fingerprint `json:"sender" gorm:"column:sender;not null"`
	CreateBulletin
}

func (Bulletin) TableName() string {
	return "bulletin_board"
}

func (b *Bulletin) Hash() string {
	hashStrings := []string{
		string(b.Sender),
		string(b.Topic),
		string(b.Content),
		b.CreatedAt.Format(time.RFC3339Nano),
	}

	idBytes := []byte{}
	for _, s := range hashStrings {
		idBytes = append(idBytes, []byte(s)...)
	}

	hash := sha256.Sum256([]byte(idBytes))

	return hex.EncodeToString(hash[:])
}

func (m *Bulletin) BeforeCreate(tx *gorm.DB) error {
	sender, recipients, encrypted, signed, err := m.Content.Analyze()
	if err != nil {
		return err
	}

	// Check for valid bulletin post content
	if !signed {
		return fmt.Errorf("bulletin post content must be signed")
	}
	if encrypted {
		return fmt.Errorf("bulletin post content must not be encrypted")
	}
	if len(recipients) > 0 {
		return fmt.Errorf("bulletin post content must not have recipients")
	}

	// Check for tampered data during synchronization
	if (m.Sender != "" && m.Sender != sender) {
		return fmt.Errorf("bulletin post sender does not match content")
	} else {
		m.Sender = sender
	}
	
	m.Base.ID = m.Hash()
	m.Base.BeforeCreate(tx)

	return nil
}