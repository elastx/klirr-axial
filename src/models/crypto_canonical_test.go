package models

import (
	"strings"
	"testing"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

func TestIsCanonicalFingerprintValid(t *testing.T) {
    // Generate a test key and derive canonical encryption fingerprint
    key, err := crypto.GenerateKey("canon", "canon@example.com", "", 0)
    if err != nil { t.Fatalf("generate key: %v", err) }
    pubArm, err := key.GetArmoredPublicKey()
    if err != nil { t.Fatalf("armored pub: %v", err) }

    pk := PublicKey(pubArm)
    canonical, err := pk.GetCanonicalEncryptionFingerprint()
    if err != nil { t.Fatalf("canonical: %v", err) }

    fp := string(canonical)
    if !IsCanonicalFingerprint(fp) {
        t.Fatalf("expected canonical fingerprint to be valid: %q", fp)
    }

    // Uppercase should be invalid (we normalize to lowercase everywhere)
    if IsCanonicalFingerprint(strings.ToUpper(fp)) {
        t.Fatalf("expected uppercase fingerprint to be invalid: %q", strings.ToUpper(fp))
    }
}

func TestIsCanonicalFingerprintRejectsNonHexOrWrongLength(t *testing.T) {
    cases := []string{
        "",                 // empty
        "12345",            // too short
        "g1234567890abcdef", // non-hex char
        "0123456789abcdef00", // too long
        "0123 456789abcdef",  // whitespace
    }
    for _, c := range cases {
        if IsCanonicalFingerprint(c) {
            t.Fatalf("expected invalid for %q", c)
        }
    }
}
