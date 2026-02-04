package models

import (
	"strings"
	"testing"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newMemDB(t *testing.T) *gorm.DB {
    t.Helper()
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil { t.Fatalf("open sqlite: %v", err) }
    if err := db.AutoMigrate(&User{}); err != nil { t.Fatalf("migrate: %v", err) }
    return db
}

func deriveExpectedEncryptionID(t *testing.T, pubArm string) string {
    t.Helper()
    key, err := crypto.NewKeyFromArmored(pubArm)
    if err != nil { t.Fatalf("read key: %v", err) }
    kr, err := crypto.NewKeyRing(key)
    if err != nil { t.Fatalf("keyring: %v", err) }
    pm := crypto.NewPlainMessage([]byte("probe"))
    enc, err := kr.Encrypt(pm, nil)
    if err != nil { t.Fatalf("encrypt: %v", err) }
    ids, _ := enc.GetHexEncryptionKeyIDs()
    if len(ids) == 0 { t.Fatalf("no encryption recipients returned") }
    return strings.ToLower(ids[0])
}

func TestUserFingerprintDerivedFromEncryptionKeyID(t *testing.T) {
    db := newMemDB(t)

    // Generate a test key
    key, err := crypto.GenerateKey("user", "user@example.com", "", 0)
    if err != nil { t.Fatalf("generate key: %v", err) }
    pubArm, err := key.GetArmoredPublicKey()
    if err != nil { t.Fatalf("armored pub: %v", err) }

    expected := deriveExpectedEncryptionID(t, pubArm)

    u := User{CreateUser: CreateUser{PublicKey: pubArm}}
    if err := db.Create(&u).Error; err != nil {
        t.Fatalf("create user: %v", err)
    }

    if strings.ToLower(u.Fingerprint) != expected {
        t.Fatalf("fingerprint not derived from encryption key ID: got %q want %q", u.Fingerprint, expected)
    }
}

func TestUserCreationRejectsSigningFingerprint(t *testing.T) {
    db := newMemDB(t)

    // Generate a test key
    key, err := crypto.GenerateKey("user2", "user2@example.com", "", 0)
    if err != nil { t.Fatalf("generate key: %v", err) }
    pubArm, err := key.GetArmoredPublicKey()
    if err != nil { t.Fatalf("armored pub: %v", err) }

    // Signing fingerprint (primary key ID) differs from encryption subkey ID
    signing := strings.ToLower(key.GetHexKeyID())
    // Sanity: expected encryption ID
    expectedEnc := deriveExpectedEncryptionID(t, pubArm)
    if signing == expectedEnc {
        t.Fatalf("test precondition failed: signing == encryption (%s)", signing)
    }

    // Attempt to create with signing fingerprint supplied should be rejected
    u := User{CreateUser: CreateUser{PublicKey: pubArm}, Fingerprint: signing}
    if err := db.Create(&u).Error; err == nil {
        t.Fatalf("expected rejection when supplied fingerprint=%q (signing), but creation succeeded; stored fp=%q expected enc=%q", signing, u.Fingerprint, expectedEnc)
    }
}
