package models

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// API create type
type CreateUser struct {
	PublicKey string `json:"public_key" gorm:"column:public_key;not null"`
}

// Stored type
type User struct {
	Base
	CreateUser
	// Canonical fingerprint: encryption subkey KeyID (16-hex, lowercase)
	Fingerprint string `json:"fingerprint" gorm:"uniqueIndex"`
}

// Gorm setup
func (User) TableName() string {
	return "users"
}

func (u *User) Hash() string {
	return u.Fingerprint
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	pubKey := u.GetPublicKey()
	encFPs, err := pubKey.GetEncryptionFingerprints()
	if err != nil {
		return err
	}
	if len(encFPs) == 0 {
		return fmt.Errorf("no encryption key id found for public key")
	}
	canonical := strings.ToLower(string(encFPs[0]))

	// Validate supplied fingerprint if present
	if u.Fingerprint != "" && strings.ToLower(u.Fingerprint) != canonical {
		return fmt.Errorf("supplied fingerprint does not match encryption key id")
	}

	u.Fingerprint = canonical

	u.Base.ID = u.Hash()
	u.Base.BeforeCreate(tx)
	return nil
}

// User methods
func (u *User) GetPublicKey() PublicKey {
	return PublicKey(u.PublicKey)
}

func (u *User) GetFingerprint() Fingerprint {
	return Fingerprint(u.Fingerprint)
}

func (u *User) SetFingerprint(fingerprint Fingerprint) {
	u.Fingerprint = string(fingerprint)
}
