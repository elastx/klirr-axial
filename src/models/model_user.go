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
	Fingerprint string    `json:"fingerprint" gorm:"uniqueIndex"`
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
	fingerprint, err := pubKey.GetFingerprint()
	if err != nil {
		return err
	}

	// Check for data manipulation during sync
	if u.Fingerprint != "" && u.GetFingerprint() != fingerprint {
		return fmt.Errorf("supplied fingerprint does not match public key")
	} else {
		u.SetFingerprint(fingerprint)
	}

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
