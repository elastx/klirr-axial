package models

import (
	"fmt"
	"strings"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

type Crypto string;

func (c* Crypto) Analyze() (Fingerprint, []Fingerprint, bool, bool, error) {
	sender := Fingerprint("")
	recipients := []Fingerprint{}
	encrypted := false
	signed := false

	// Check if it's a valid PGP message
	message, err := crypto.NewPGPMessageFromArmored(string(*c))
	if err != nil || message == nil {
		// Attempt to parse clearsigned message
		content := string(*c)
		if strings.Contains(content, "-----BEGIN PGP SIGNED MESSAGE-----") &&
		   strings.Contains(content, "-----BEGIN PGP SIGNATURE-----") &&
		   strings.Contains(content, "-----END PGP SIGNATURE-----") {
			// Extract signature block
			sigStart := strings.Index(content, "-----BEGIN PGP SIGNATURE-----")
			sigEnd := strings.Index(content, "-----END PGP SIGNATURE-----")
			if sigStart >= 0 && sigEnd > sigStart {
				sigEnd += len("-----END PGP SIGNATURE-----")
				sigArmored := content[sigStart:sigEnd]
				signature := Signature(sigArmored)
				signer, sigErr := signature.GetSignerFingerprint()
				if sigErr != nil {
					return sender, recipients, encrypted, signed, fmt.Errorf("invalid clearsigned PGP signature: %w", sigErr)
				}
				// For clearsigned messages: signed=true, no recipients, not encrypted
				signed = true
				encrypted = false
				sender = signer
				return sender, recipients, encrypted, signed, nil
			}
		}
		return sender, recipients, encrypted, signed, fmt.Errorf("invalid PGP message: %w", err)
	}
	
	recipientStrings, _ := message.GetHexEncryptionKeyIDs()
	for _, r := range recipientStrings {
		if r == "" { continue }
		recipients = append(recipients, Fingerprint(strings.ToLower(r)))
	}
	if len(recipients) > 0 {
		encrypted = true
	}

	senderStrings, _ := message.GetHexSignatureKeyIDs()
	if len(senderStrings) > 0 {
		s := senderStrings[0]
		if s != "" { sender = Fingerprint(strings.ToLower(s)) }
		signed = true
	}

	return sender, recipients, encrypted, signed, nil
}

