package models

import (
	"fmt"

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
	// Primary fingerprint retained for compatibility; represents canonical signing key ID
	Fingerprint string    `json:"fingerprint" gorm:"uniqueIndex"`
	// Explicit signing key ID (often same as primary)
	SigningFingerprint string `json:"signing_fingerprint" gorm:"column:signing_fingerprint;index"`
	// All encryption subkey IDs for this user
	EncryptionFingerprints Fingerprints `json:"encryption_fingerprints,omitempty" gorm:"column:encryption_fingerprints;type:jsonb"`
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
	// Derive signing and encryption fingerprints from the supplied public key
	signingFP, err := pubKey.GetSigningFingerprint()
	if err != nil {
		return err
	}
	encFPs, err := pubKey.GetEncryptionFingerprints()
	if err != nil {
		return err
	}

	// Check for data manipulation during sync
	if u.Fingerprint != "" && u.GetFingerprint() != signingFP {
		return fmt.Errorf("supplied fingerprint does not match public key")
	} else {
		u.SetFingerprint(signingFP)
	}

	// Populate extended fingerprint fields
	u.SigningFingerprint = string(signingFP)
	u.EncryptionFingerprints = encFPs

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
