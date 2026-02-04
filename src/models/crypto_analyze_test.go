package models

import (
	"strings"
	"testing"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

func TestAnalyzeLowercaseRecipients(t *testing.T) {
    // Generate a test key
    key, err := crypto.GenerateKey("analyze", "analyze@example.com", "", 0)
    if err != nil { t.Fatalf("generate key: %v", err) }

    kr, err := crypto.NewKeyRing(key)
    if err != nil { t.Fatalf("keyring: %v", err) }

    pm := crypto.NewPlainMessage([]byte("probe"))
    enc, err := kr.Encrypt(pm, nil)
    if err != nil { t.Fatalf("encrypt: %v", err) }

    armored, err := enc.GetArmored()
    if err != nil { t.Fatalf("armored: %v", err) }

    c := Crypto(armored)
    _, recipients, encrypted, _, err := c.Analyze()
    if err != nil { t.Fatalf("analyze: %v", err) }
    if !encrypted { t.Fatalf("expected encrypted=true") }
    if len(recipients) == 0 { t.Fatalf("expected recipients > 0") }

    for _, fp := range recipients {
        got := string(fp)
        if got != strings.ToLower(got) {
            t.Fatalf("recipient fingerprint not lowercase: %q", got)
        }
    }
}
