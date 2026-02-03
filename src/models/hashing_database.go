package models

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type HashSet struct {
	Messages string `json:"messages"`
	Users    string `json:"users"`
	Full     string `json:"full"`
}

var (
	hashes = HashSet{
		Messages: "",
		Users:    "",
		Full:     "",
	}
)

// GetMessagesHash calculates a hash of message IDs ordered by timestamp
// If timeRange is provided, only messages within that range are included
func GetMessagesHash(db *gorm.DB, start, end *time.Time) (string, error) {
	query := db.Model(&Message{}).Order("created_at")

	if start != nil {
		query = query.Where("created_at >= ?", start)
	}
	if end != nil {
		query = query.Where("created_at <= ?", end)
	}

	var messageIDs []string
	// Prefer current schema `id`; fall back to legacy `message_id` only if present
	if err := query.Select("id AS combined_id").Pluck("combined_id", &messageIDs).Error; err != nil {
		// If selecting `id` fails for any reason, try legacy column (older deployments)
		// Note: referencing a missing column causes SQLSTATE 42703; guard by checking error text
		legacyErr := query.Select("message_id AS combined_id").Pluck("combined_id", &messageIDs).Error
		if legacyErr != nil {
			// Provide clearer context, include both errors
			combined := fmt.Errorf("failed to get message IDs (id err: %v, legacy err: %v)", err, legacyErr)
			// If error mentions missing column, surface a concise hint
			if strings.Contains(strings.ToLower(legacyErr.Error()), "does not exist") || strings.Contains(legacyErr.Error(), "42703") {
				return "", fmt.Errorf("failed to get message IDs: column 'message_id' not found; current schema uses 'id': %w", combined)
			}
			return "", combined
		}
	}

	hasher := sha256.New()
	for _, id := range messageIDs {
		hasher.Write([]byte(id))
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// GetUsersHash calculates a hash of user fingerprints in alphabetical order
func GetUsersHash(db *gorm.DB) (string, error) {
	var fingerprints []string
	if err := db.Model(&User{}).Order("fingerprint").Pluck("fingerprint", &fingerprints).Error; err != nil {
		return "", fmt.Errorf("failed to get user fingerprints: %v", err)
	}

	hasher := sha256.New()
	for _, fp := range fingerprints {
		hasher.Write([]byte(fp))
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func RefreshHashes(db *gorm.DB) error {
	messagesHash, err := GetMessagesHash(db, nil, nil)
	if err != nil {
		return err
	}

	usersHash, err := GetUsersHash(db)
	if err != nil {
		return err
	}

	// Combine hashes in a deterministic order
	hasher := sha256.New()
	hasher.Write([]byte("messages:" + messagesHash))
	hasher.Write([]byte("users:" + usersHash))

	hashes = HashSet{
		Messages: messagesHash,
		Users:    usersHash,
		Full:     hex.EncodeToString(hasher.Sum(nil)),
	}

	UpdateHashes(hashes)

	return nil
}

// GetDatabaseHashes returns a hash combining all table hashes
func GetDatabaseHashes(db *gorm.DB) (HashSet, error) {
	if hashes.Full == "" {
		err := RefreshHashes(db)
		if err != nil {
			return HashSet{}, err
		}
	}
	return hashes, nil
}

// GetMessageHashForRange returns a hash of messages within a specific time range
func GetMessageHashForRange(db *gorm.DB, start, end time.Time) (string, error) {
	return GetMessagesHash(db, &start, &end)
}

func GetUsersHashByFingerprintRange(db *gorm.DB, start, end string) (string, error) {
	query := db.Model(&User{}).Where("fingerprint >= ?", start).Where("fingerprint <= ?", end)

	var fingerprints []string
	if err := query.Order("fingerprint").Pluck("fingerprint", &fingerprints).Error; err != nil {
		return "", fmt.Errorf("failed to get user fingerprints: %v", err)
	}

	hasher := sha256.New()
	for _, fp := range fingerprints {
		hasher.Write([]byte(fp))
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}