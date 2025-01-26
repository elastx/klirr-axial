package models

import (
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
)

type RemoteNode struct {
	Hash string `json:"hash"`
	Address string `json:"base_url"`
}

// SyncState manages the synchronization state
type SyncState struct {
	mu        sync.RWMutex
	isSyncing bool
	hash      string
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
	Hash  string     `json:"hash"`
}

type MessagesPeriod struct {
	Period
	Messages []Message `json:"messages"`
}

// type UsersRange struct {
// 	Start string `json:"start"`
// 	End   string `json:"end"`
// }

// type HashedUsersRange struct {
// 	UsersRange
// 	Hash  string `json:"hash"`
// }

// type ListUsersRange struct {
// 	UsersRange
// 	Users  []User `json:"users"`
// }

// SyncRequest represents a request to sync data
type SyncRequest struct {
	Ranges []HashedPeriod `json:"ranges"`
}

// SyncResponse can either contain data or more ranges to check
type SyncResponse struct {
	Hash     string         `json:"hash"`
	IsBusy   bool           `json:"is_busy"`
	Ranges   []HashedPeriod `json:"ranges,omitempty"`
	Messages []MessagesPeriod `json:"messages,omitempty"`
	// Users    []User         `json:"users,omitempty"`
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

// UpdateHash updates the current database hash
func UpdateHash(hash string) {
	syncState.mu.Lock()
	defer syncState.mu.Unlock()
	syncState.hash = hash
}

// GetHash returns the current database hash
func GetHash() string {
	syncState.mu.RLock()
	defer syncState.mu.RUnlock()
	return syncState.hash
}

// GenerateHashRanges creates the standard set of time ranges to check
func GenerateHashRanges(db *gorm.DB, periods []Period) ([]HashedPeriod, error) {
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
