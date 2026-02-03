package models

import (
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