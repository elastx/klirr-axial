package models

import (
	"strings"
	"testing"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

func TestGetSignerFingerprintLowercase(t *testing.T) {
    // Generate a test key
    key, err := crypto.GenerateKey("signer", "signer@example.com", "", 0)
    if err != nil { t.Fatalf("generate key: %v", err) }

    kr, err := crypto.NewKeyRing(key)
    if err != nil { t.Fatalf("keyring: %v", err) }

    pm := crypto.NewPlainMessage([]byte("hello"))

    // Create a detached signature and armor it
    sigObj, err := kr.SignDetached(pm)
    if err != nil { t.Fatalf("sign detached: %v", err) }
    sigArmored, err := sigObj.GetArmored()
    if err != nil { t.Fatalf("armored signature: %v", err) }

    sig := Signature(sigArmored)
    signer, err := sig.GetSignerFingerprint()
    if err != nil { t.Fatalf("get signer: %v", err) }

    got := string(signer)
    if got != strings.ToLower(got) {
        t.Fatalf("signer fingerprint not lowercase: %q", got)
    }
}
