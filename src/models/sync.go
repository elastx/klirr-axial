package models

import (
	"fmt"
	"sync"

	"axial/hashrange"

	"gorm.io/gorm"
)

// SyncState manages the synchronization state
type SyncState struct {
	mu        sync.RWMutex
	isSyncing bool
	hashes    HashSet
}

var (
	syncState = &SyncState{}
)

// MessagesPeriod carries the Messages that live within a hash range. The
// Period bounds are inlined into JSON via the embedded hashrange.Period.
type MessagesPeriod struct {
	hashrange.Period
	Messages []Message `json:"messages"`
}

// BulletinsPeriod carries the Bulletins that live within a hash range.
type BulletinsPeriod struct {
	hashrange.Period
	Bulletins []Bulletin `json:"bulletins"`
}

// UsersRange carries the Users that live within a fingerprint hash range.
type UsersRange struct {
	hashrange.StringRange
	Users []User `json:"users"`
}

// StartSync attempts to start a sync operation
func StartSync() bool {
	syncState.mu.Lock()
	defer syncState.mu.Unlock()

	if syncState.isSyncing {
		return false
	}

	syncState.isSyncing = true
	return true
}

// EndSync marks the sync operation as complete
func EndSync() {
	syncState.mu.Lock()
	defer syncState.mu.Unlock()
	syncState.isSyncing = false
}

// IsSyncing checks if a sync is in progress
func IsSyncing() bool {
	syncState.mu.RLock()
	defer syncState.mu.RUnlock()
	return syncState.isSyncing
}

// UpdateHashes updates the current database hash
func UpdateHashes(hash HashSet) {
	syncState.mu.Lock()
	defer syncState.mu.Unlock()
	syncState.hashes = hashes
}

// GetHashes returns the current database hash
func GetHashes() HashSet {
	syncState.mu.RLock()
	defer syncState.mu.RUnlock()
	return syncState.hashes
}

// GetMessagesHashRanges computes the message-content hash for each of the
// supplied hash ranges. Pure-math splitting / comparison of ranges lives in
// the hashrange package; this function is the DB-touching adapter.
func GetMessagesHashRanges(db *gorm.DB, periods []hashrange.Period) ([]hashrange.HashedPeriod, error) {
	hashedPeriods := []hashrange.HashedPeriod{}
	for _, period := range periods {
		hash, err := GetMessagesHash(db, period.Start, period.End)
		if err != nil {
			return nil, fmt.Errorf("failed to get messages hash: %v", err)
		}
		hashedPeriods = append(hashedPeriods, hashrange.HashedPeriod{
			Period: period,
			Hash:   hash,
		})
	}
	return hashedPeriods, nil
}

// GetBulletinsHashRanges computes the bulletin-content hash for each of the
// supplied hash ranges.
func GetBulletinsHashRanges(db *gorm.DB, periods []hashrange.Period) ([]hashrange.HashedPeriod, error) {
	hashedPeriods := []hashrange.HashedPeriod{}
	for _, period := range periods {
		hash, err := GetBulletinsHash(db, period.Start, period.End)
		if err != nil {
			return nil, fmt.Errorf("failed to get bulletins hash: %v", err)
		}
		hashedPeriods = append(hashedPeriods, hashrange.HashedPeriod{
			Period: period,
			Hash:   hash,
		})
	}
	return hashedPeriods, nil
}

// GetUsersHashRanges computes the user-content hash for each of the supplied
// fingerprint hash ranges.
func GetUsersHashRanges(db *gorm.DB, stringRanges []hashrange.StringRange) ([]hashrange.HashedUsersRange, error) {
	hashedRanges := []hashrange.HashedUsersRange{}
	for _, stringRange := range stringRanges {
		hash, err := GetUsersHashByFingerprintRange(db, stringRange.Start, stringRange.End)
		if err != nil {
			return nil, fmt.Errorf("failed to get users hash: %v", err)
		}
		hashedRanges = append(hashedRanges, hashrange.HashedUsersRange{
			StringRange: stringRange,
			Hash:        hash,
		})
	}

	return hashedRanges, nil
}

func GetMessagesByPeriod(db *gorm.DB, period hashrange.Period) ([]Message, error) {
	var messages []Message
	err := db.Where("created_at >= ? AND created_at < ?", period.Start, period.End).Find(&messages).Error
	return messages, err
}

func CountMessagesByPeriod(db *gorm.DB, period hashrange.Period) int64 {
	var count int64
	db.Model(&Message{}).Where("created_at >= ? AND created_at < ?", period.Start, period.End).Count(&count)
	return count
}

func GetBulletinsByPeriod(db *gorm.DB, period hashrange.Period) ([]Bulletin, error) {
	var bulletins []Bulletin
	err := db.Where("created_at >= ? AND created_at < ?", period.Start, period.End).Find(&bulletins).Error
	return bulletins, err
}

func CountBulletinsByPeriod(db *gorm.DB, period hashrange.Period) int64 {
	var count int64
	db.Model(&Bulletin{}).Where("created_at >= ? AND created_at < ?", period.Start, period.End).Count(&count)
	return count
}

func GetUsersByFingerprintRange(db *gorm.DB, start, end string) ([]User, error) {
	var users []User
	err := db.Where("fingerprint >= ? AND fingerprint < ?", start, end).Order("fingerprint").Find(&users).Error
	return users, err
}

func CountUsersByFingerprintRange(db *gorm.DB, start, end string) int64 {
	var count int64
	db.Model(&User{}).Where("fingerprint >= ? AND fingerprint < ?", start, end).Count(&count)
	return count
}
