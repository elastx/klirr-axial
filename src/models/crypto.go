package models

import (
	"fmt"
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

