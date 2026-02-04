package models

import (
	"fmt"
	"strings"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

type Signature string;

func (s* Signature) PGP() (*crypto.PGPSignature, error) {
	return crypto.NewPGPSignatureFromArmored(string(*s))
}

func (s* Signature) GetSignerFingerprint() (Fingerprint, error) {
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

	return Fingerprint(strings.ToLower(signer[0])), nil
}

func (s* Signature) Verify(publicKey PublicKey) error {
	// Check if it's a valid PGP signature
	signature, err := s.PGP()
	if err != nil {
		return err
	}

	// Check if the public key is valid
	key, err := publicKey.PGP()
	if err != nil {
		return err
	}

	// Get key from signature
	signatureKey, ok := signature.GetHexSignatureKeyIDs()
	if !ok {
		return fmt.Errorf("signature does not contain a key ID")
	}
	
	// Check that there's only one key ID
	if len(signatureKey) != 1 {
		return fmt.Errorf("signature must have exactly one key ID")
	}

	// Check if the key ID matches the public key
	if signatureKey[0] != key.GetFingerprint() {
		return fmt.Errorf("signature key ID does not match public key")
	}

	return nil
}