package models

import (
	"strings"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

type PublicKey string;

func (pk* PublicKey) PGP() (*crypto.Key, error) {
	return crypto.NewKeyFromArmored(string(*pk))
}

func (pk* PublicKey) GetFingerprint() (Fingerprint, error) {
	key, err := pk.PGP()
	if err != nil {
		return "", err
	}

	return Fingerprint(key.GetHexKeyID()), nil
}

// GetSigningFingerprint returns the canonical signing key ID derived from the public key.
// For single-key setups, this is equivalent to the primary key ID.
func (pk* PublicKey) GetSigningFingerprint() (Fingerprint, error) {
	return pk.GetFingerprint()
}

// GetEncryptionFingerprints derives the encryption subkey IDs by encrypting a probe
// message to this public key and inspecting the resulting recipients.
func (pk* PublicKey) GetEncryptionFingerprints() (Fingerprints, error) {
	key, err := pk.PGP()
	if err != nil {
		return nil, err
	}

	// Build a keyring for encryption
	kr, err := crypto.NewKeyRing(key)
	if err != nil {
		return nil, err
	}

	// Encrypt a small probe message to collect recipient key IDs
	pm := crypto.NewPlainMessage([]byte("probe"))
	encMsg, err := kr.Encrypt(pm, nil)
	if err != nil {
		return nil, err
	}

	// Extract encryption key IDs (hex)
	ids, _ := encMsg.GetHexEncryptionKeyIDs()
	fps := make(Fingerprints, 0, len(ids))
	for _, id := range ids {
		if id != "" {
			fps = append(fps, Fingerprint(id))
		}
	}
	return fps, nil
}

// GetCanonicalEncryptionFingerprint selects the canonical encryption key ID.
// Preference order:
// 1. First encryption recipient key ID returned by gopenpgp (stable order)
// 2. Fallback to the primary key ID if encryption IDs are unavailable
// The returned value is normalized to lowercase.
func (pk* PublicKey) GetCanonicalEncryptionFingerprint() (Fingerprint, error) {
	key, err := pk.PGP()
	if err != nil {
		return "", err
	}

	// Try encryption recipients
	kr, err := crypto.NewKeyRing(key)
	if err != nil {
		return "", err
	}
	pm := crypto.NewPlainMessage([]byte("probe"))
	encMsg, err := kr.Encrypt(pm, nil)
	if err == nil {
		ids, _ := encMsg.GetHexEncryptionKeyIDs()
		if len(ids) > 0 && ids[0] != "" {
			return Fingerprint(strings.ToLower(ids[0])), nil
		}
	}

	// Fallback to primary key ID
	return Fingerprint(strings.ToLower(key.GetHexKeyID())), nil
}