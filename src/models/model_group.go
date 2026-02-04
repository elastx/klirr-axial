package models

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// API create type
type CreateGroup struct {
	UserID string `json:"user_id" gorm:"column:user_id;not null;uniqueIndex:idx_group_user"`
	PrivateKey Crypto `json:"private_key" gorm:"column:private_key;not null"`
}

// Stored type
type Group struct {
	Base
	CreateGroup
	Members Fingerprints `json:"members,omitempty" gorm:"column:members;type:jsonb"`
}

type HydratedGroup struct {
	Group
	User User `json:"user" gorm:"foreignKey:UserID;references:Fingerprint"`
	Users []User `json:"users" gorm:"-"`
}

// Gorm setup
func (Group) TableName() string {
	return "groups"
}

func (g *Group) Hash() string {
		// Create a string combining all relevant properties
	recipients := ""
	for _, r := range g.Members {
		recipients += string(r)
	}
	hashStrings := []string{
		g.UserID,
		string(g.PrivateKey),
		recipients,
		g.CreatedAt.Format(time.RFC3339Nano),
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

func (g *Group) BeforeCreate(tx *gorm.DB) error {

	privKeyMessage := g.GetPrivateKey()
	_, members, encrypted, signed, err := privKeyMessage.Analyze()
	if err != nil {
		return err
	}

	if !encrypted || signed {
		return fmt.Errorf("invalid group private key")
	}

	g.Members = members

	g.Base.ID = g.Hash()
	g.Base.BeforeCreate(tx)
	return nil
}

// Group methods
func (g *Group) GetPrivateKey() Crypto {
	return Crypto(g.PrivateKey)
}
