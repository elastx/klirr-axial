package models

import (
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
)

// SyncState manages the synchronization state
type SyncState struct {
	mu        sync.RWMutex
	isSyncing bool
	hashes      HashSet
}

var (
	syncState = &SyncState{}
)

type Period struct {
	Start *time.Time `json:"start,omitempty"`
	End   *time.Time `json:"end,omitempty"`
}

func RealizeStart(start *time.Time) time.Time {
	if start == nil {
		// Start at 2025-01-01, the release year
		return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	return *start
}

func RealizeEnd(end *time.Time) time.Time {
	if end == nil {
		// End at the current time
		return time.Now()
	}
	return *end
}

// HashedPeriod represents a time period and its hash
type HashedPeriod struct {
	Period
	Hash string `json:"hash"`
}

type MessagesPeriod struct {
	Period
	Messages []Message `json:"messages"`
}

type BulletinsPeriod struct {
	Period
	Bulletins []Bulletin `json:"bulletins"`
}

type StringRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type UsersRange struct {
	StringRange
	Users  []User `json:"users"`
}

type HashedUsersRange struct {
	StringRange
	Hash  string `json:"hash"`
}

type ListUsersRange struct {
	StringRange
	Users  []User `json:"users"`
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

// GetMessagesHashRanges creates the standard set of time ranges to check
func GetMessagesHashRanges(db *gorm.DB, periods []Period) ([]HashedPeriod, error) {
	hashedPeriods := []HashedPeriod{}
	for _, period := range periods {
		hash, err := GetMessagesHash(db, period.Start, period.End)
		if err != nil {
			return nil, fmt.Errorf("failed to get messages hash: %v", err)
		}
		hashedPeriods = append(hashedPeriods, HashedPeriod{
			Period: period,
			Hash:   hash,
		})
	}
	return hashedPeriods, nil
}

// GetBulletinsHashRanges creates hashed ranges for bulletins
func GetBulletinsHashRanges(db *gorm.DB, periods []Period) ([]HashedPeriod, error) {
	hashedPeriods := []HashedPeriod{}
	for _, period := range periods {
		hash, err := GetBulletinsHash(db, period.Start, period.End)
		if err != nil {
			return nil, fmt.Errorf("failed to get bulletins hash: %v", err)
		}
		hashedPeriods = append(hashedPeriods, HashedPeriod{
			Period: period,
			Hash:   hash,
		})
	}
	return hashedPeriods, nil
}

func GetUsersHashRanges(db *gorm.DB, stringRanges []StringRange) ([]HashedUsersRange, error) {
	hashedRanges := []HashedUsersRange{}
	for _, stringRange := range stringRanges {
		hash, err := GetUsersHashByFingerprintRange(db, stringRange.Start, stringRange.End)
		if err != nil {
			return nil, fmt.Errorf("failed to get users hash: %v", err)
		}
		hashedRanges = append(hashedRanges, HashedUsersRange{
			StringRange: stringRange,
			Hash:        hash,
		})
	}

	return hashedRanges, nil
}

// SplitTimeRange splits a time range into n equal parts
func SplitTimeRange(period Period, n int) []Period {
	start := RealizeStart(period.Start)
	end := RealizeEnd(period.End)

	duration := end.Sub(start)
	partDuration := duration / time.Duration(n)

	ranges := make([]Period, n)
	for i := 0; i < n; i++ {
		partStart := start.Add(partDuration * time.Duration(i))
		partEnd := partStart.Add(partDuration)
		if i == n-1 {
			partEnd = end // Ensure we don't miss any time due to rounding
		}

		ranges[i] = Period{
			Start: &partStart,
			End:   &partEnd,
		}
	}

	return ranges
}

func GetMessagesByPeriod(db *gorm.DB, period Period) ([]Message, error) {
	var messages []Message
	err := db.Where("created_at >= ? AND created_at < ?", period.Start, period.End).Find(&messages).Error
	return messages, err
}

func CountMessagesByPeriod(db *gorm.DB, period Period) int64 {
	var count int64
	db.Model(&Message{}).Where("created_at >= ? AND created_at < ?", period.Start, period.End).Count(&count)
	return count
}


func GetBulletinsByPeriod(db *gorm.DB, period Period) ([]Bulletin, error) {
	var bulletins []Bulletin
	err := db.Where("created_at >= ? AND created_at < ?", period.Start, period.End).Find(&bulletins).Error
	return bulletins, err
}

func CountBulletinsByPeriod(db *gorm.DB, period Period) int64 {
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