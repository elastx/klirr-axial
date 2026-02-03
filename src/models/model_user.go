package models

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

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
	Fingerprint string `json:"fingerprint" gorm:"uniqueIndex"`
	Signature   string      `json:"signature,omitempty"`
	Signer      string `json:"signer,omitempty"`
	SignedAt    time.Time   `json:"signed_at,omitempty"`
}

// Gorm setup
func (User) TableName() string {
	return "users"
}

func (u *User) Hash() string {
	idStrings := []string{
		string(u.Fingerprint),
		string(u.Signer),
		string(u.Signature),
		u.SignedAt.Format(time.RFC3339Nano),
		u.CreatedAt.Format(time.RFC3339Nano),
	}

	idString := strings.Join(idStrings, "|")

	hash := sha256.Sum256([]byte(idString))

	return hex.EncodeToString(hash[:])
}

func (u *User) BeforeCreate(*gorm.DB) error {

	pubKey := u.GetPublicKey()
	fingerprint, err := pubKey.GetFingerprint()
	if err != nil {
		return err
	}
	
	if u.Signature != "" {
		signature := Signature(u.Signature)
		signer, err := signature.GetSignerFingerprint()
		if err != nil {
			return err
		}

		// Check for data manipulation during sync
		if u.Signer != "" && signer != u.GetSigner() {
			return fmt.Errorf("signature does not match signer")
		} else {
			u.SetSigner(signer)
		}
	}

	// Check for data manipulation during sync
	if u.Fingerprint != "" && u.GetFingerprint() != fingerprint {
		return fmt.Errorf("supplied fingerprint does not match public key")
	} else {
		u.SetFingerprint(fingerprint)
	}

	u.Base.BeforeCreate(u.Hash())
	return nil
}

// User methods
func (u *User) GetPublicKey() PublicKey {
	return PublicKey(u.PublicKey)
}

func (u *User) GetSignature() Signature {
	return Signature(u.Signature)
}

func (u *User) GetFingerprint() Fingerprint {
	return Fingerprint(u.Fingerprint)
}

func (u *User) SetFingerprint(fingerprint Fingerprint) {
	u.Fingerprint = string(fingerprint)
} 

func (u *User) GetSigner() Fingerprint {
	return Fingerprint(u.Signer)
}

func (u *User) SetSigner(signer Fingerprint) {
	u.Signer = string(signer)
}
