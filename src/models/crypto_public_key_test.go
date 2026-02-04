package models

import (
	"strings"
	"testing"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

func TestCanonicalEncryptionFingerprintLowercase(t *testing.T) {
    // Generate a test key
    key, err := crypto.GenerateKey("lowercase", "lower@example.com", "", 0)
    if err != nil { t.Fatalf("generate key: %v", err) }
    pubArm, err := key.GetArmoredPublicKey()
    if err != nil { t.Fatalf("armored pub: %v", err) }

    pk := PublicKey(pubArm)

    canonical, err := pk.GetCanonicalEncryptionFingerprint()
    if err != nil { t.Fatalf("canonical encryption fingerprint: %v", err) }

    // Derive expected via encryption recipients
    kr, err := crypto.NewKeyRing(key)
    if err != nil { t.Fatalf("keyring: %v", err) }
    pm := crypto.NewPlainMessage([]byte("probe"))
    enc, err := kr.Encrypt(pm, nil)
    if err != nil { t.Fatalf("encrypt: %v", err) }
    ids, _ := enc.GetHexEncryptionKeyIDs()
    if len(ids) == 0 { t.Fatalf("no encryption recipients returned") }

    expected := strings.ToLower(ids[0])
    got := string(canonical)

    if got != expected {
        t.Fatalf("canonical encryption fingerprint mismatch: got %q want %q", got, expected)
    }

    // Ensure lowercase normalization
    if got != strings.ToLower(got) {
        t.Fatalf("canonical fingerprint not lowercase: %q", got)
    }
}

func TestCanonicalEncryptionFingerprintInvalidKeyErrors(t *testing.T) {
    pk := PublicKey("-----BEGIN PGP PUBLIC KEY BLOCK-----\ninvalid\n-----END PGP PUBLIC KEY BLOCK-----")
    if _, err := pk.GetCanonicalEncryptionFingerprint(); err == nil {
        t.Fatalf("expected error for invalid armored public key")
    }
}
