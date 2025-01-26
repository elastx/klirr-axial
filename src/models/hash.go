package models

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"gorm.io/gorm"
)

var (
	hash = ""
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
	if err := query.Pluck("message_id", &messageIDs).Error; err != nil {
		return "", fmt.Errorf("failed to get message IDs: %v", err)
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

func RefreshHash(db *gorm.DB) error {
	messagesHash, err := GetMessagesHash(db, nil, nil)
	if err != nil {
		return err
	}

	// Combine hashes in a deterministic order
	hasher := sha256.New()
	hasher.Write([]byte("messages:" + messagesHash))
	// hasher.Write([]byte("users:" + usersHash))

	hash = hex.EncodeToString(hasher.Sum(nil))
	return nil
}

// GetDatabaseHash returns a hash combining all table hashes
func GetDatabaseHash(db *gorm.DB) (string, error) {
	if hash == "" {
		err := RefreshHash(db)
		if err != nil {
			return "", err
		}
	}
	return hash, nil
}

// GetMessageHashForRange returns a hash of messages within a specific time range
func GetMessageHashForRange(db *gorm.DB, start, end time.Time) (string, error) {
	return GetMessagesHash(db, &start, &end)
}
