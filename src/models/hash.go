package models

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"gorm.io/gorm"
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

// GetDatabaseHash returns a hash combining all table hashes
func GetDatabaseHash(db *gorm.DB) (string, error) {
	messagesHash, err := GetMessagesHash(db, nil, nil)
	if err != nil {
		return "", err
	}

	usersHash, err := GetUsersHash(db)
	if err != nil {
		return "", err
	}

	// Combine hashes in a deterministic order
	hasher := sha256.New()
	hasher.Write([]byte("messages:" + messagesHash))
	hasher.Write([]byte("users:" + usersHash))

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// GetMessageHashForRange returns a hash of messages within a specific time range
func GetMessageHashForRange(db *gorm.DB, start, end time.Time) (string, error) {
	return GetMessagesHash(db, &start, &end)
}

// GetHashDifference returns information about how our dataset differs from another node
func GetHashDifference(db *gorm.DB, theirHash string) (string, error) {
	ourHash, err := GetDatabaseHash(db)
	if err != nil {
		return "", fmt.Errorf("failed to get our hash: %v", err)
	}

	if ourHash == theirHash {
		return "in_sync", nil
	}

	// Get counts and latest timestamps
	var messageCount int64
	var lastMessageTime *time.Time
	var userCount int64

	if err := db.Model(&Message{}).Count(&messageCount).Error; err != nil {
		return "", fmt.Errorf("failed to count messages: %v", err)
	}

	if messageCount > 0 {
		var lastMessage Message
		if err := db.Order("created_at DESC").First(&lastMessage).Error; err != nil {
			return "", fmt.Errorf("failed to get last message: %v", err)
		}
		lastMessageTime = &lastMessage.CreatedAt
	}

	if err := db.Model(&User{}).Count(&userCount).Error; err != nil {
		return "", fmt.Errorf("failed to count users: %v", err)
	}

	diff := fmt.Sprintf("messages=%d", messageCount)
	if lastMessageTime != nil {
		diff += fmt.Sprintf(":last_message=%s", lastMessageTime.UTC().Format(time.RFC3339))
	}
	diff += fmt.Sprintf(":users=%d", userCount)

	return fmt.Sprintf("different:%s", diff), nil
} 