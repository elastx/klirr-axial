package models

import (
	"fmt"
	"strings"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

// PGP domain types. All five are string-aliased identifiers or armored PGP
// blobs that carry meaning derived from their content. The GORM Scanner/Valuer
// adapters for Fingerprint and Fingerprints live in crypto_persistence.go.

type Crypto string

type Fingerprint string

type Fingerprints []Fingerprint

type PublicKey string

type Signature string

// --- PublicKey ---------------------------------------------------------------

func (pk *PublicKey) PGP() (*crypto.Key, error) {
	return crypto.NewKeyFromArmored(string(*pk))
}

func (pk *PublicKey) GetFingerprint() (Fingerprint, error) {
	key, err := pk.PGP()
	if err != nil {
		return "", err
	}

	return Fingerprint(key.GetHexKeyID()), nil
}

// --- Signature ---------------------------------------------------------------

func (s *Signature) PGP() (*crypto.PGPSignature, error) {
	return crypto.NewPGPSignatureFromArmored(string(*s))
}

func (s *Signature) GetSignerFingerprint() (Fingerprint, error) {
	signature, err := s.PGP()
	if err != nil {
		return "", err
	}

	signer, ok := signature.GetHexSignatureKeyIDs()
	if !ok {
		return "", fmt.Errorf("signature does not contain a key ID")
	}

	if len(signer) != 1 {
		return "", fmt.Errorf("signature must have exactly one key ID")
	}

	return Fingerprint(signer[0]), nil
}

func (s *Signature) Verify(publicKey PublicKey) error {
	signature, err := s.PGP()
	if err != nil {
		return err
	}

	key, err := publicKey.PGP()
	if err != nil {
		return err
	}

	signatureKey, ok := signature.GetHexSignatureKeyIDs()
	if !ok {
		return fmt.Errorf("signature does not contain a key ID")
	}

	if len(signatureKey) != 1 {
		return fmt.Errorf("signature must have exactly one key ID")
	}

	if signatureKey[0] != key.GetFingerprint() {
		return fmt.Errorf("signature key ID does not match public key")
	}

	return nil
}

// --- Crypto (armored PGP message analyzer) -----------------------------------

func (c *Crypto) Analyze() (Fingerprint, []Fingerprint, bool, bool, error) {
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
		recipients = append(recipients, Fingerprint(r))
	}
	if len(recipients) > 0 {
		encrypted = true
	}

	senderStrings, _ := message.GetHexSignatureKeyIDs()
	if len(senderStrings) > 0 {
		sender = Fingerprint(senderStrings[0])
		signed = true
	}

	return sender, recipients, encrypted, signed, nil
}
