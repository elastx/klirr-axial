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
	hash      string
}

var (
	syncState = &SyncState{}
)

// HashedPeriod represents a time period and its hash
type HashedPeriod struct {
	Start *time.Time `json:"start,omitempty"`
	End   *time.Time `json:"end,omitempty"`
	Hash  string     `json:"hash"`
}

// SyncRequest represents a request to sync data
type SyncRequest struct {
	NodeID string         `json:"node_id"`
	Hash   string         `json:"hash"`
	Ranges []HashedPeriod `json:"ranges"`
}

// SyncResponse can either contain data or more ranges to check
type SyncResponse struct {
	NodeID   string         `json:"node_id"`
	Hash     string         `json:"hash"`
	IsBusy   bool           `json:"is_busy"`
	Ranges   []HashedPeriod `json:"ranges,omitempty"`
	Users    []User         `json:"users,omitempty"`
	Messages []Message      `json:"messages,omitempty"`
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

// getWeekStart returns the start of the week (Monday) for a given time
func getWeekStart(t time.Time) time.Time {
	weekday := t.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	return t.AddDate(0, 0, -int(weekday-time.Monday)).Truncate(24 * time.Hour)
}

// GenerateHashRanges creates the standard set of time ranges to check
func GenerateHashRanges(db *gorm.DB) ([]HashedPeriod, error) {
	now := time.Now().UTC()

	// Calculate period boundaries
	thisWeekStart := getWeekStart(now)
	lastWeekStart := thisWeekStart.AddDate(0, 0, -7)
	fourWeeksStart := lastWeekStart.AddDate(0, 0, -28)
	sixMonthsStart := lastWeekStart.AddDate(0, -6, 0)

	// Get users hash
	usersHash, err := GetUsersHash(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get users hash: %v", err)
	}

	// Get message hashes for each period
	currentWeekHash, err := GetMessagesHash(db, &thisWeekStart, &now)
	if err != nil {
		return nil, fmt.Errorf("failed to get current week hash: %v", err)
	}

	lastWeekHash, err := GetMessagesHash(db, &lastWeekStart, &thisWeekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to get last week hash: %v", err)
	}

	fourWeeksHash, err := GetMessagesHash(db, &fourWeeksStart, &lastWeekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to get four weeks hash: %v", err)
	}

	sixMonthsHash, err := GetMessagesHash(db, &sixMonthsStart, &fourWeeksStart)
	if err != nil {
		return nil, fmt.Errorf("failed to get six months hash: %v", err)
	}

	oldestHash, err := GetMessagesHash(db, nil, &sixMonthsStart)
	if err != nil {
		return nil, fmt.Errorf("failed to get oldest messages hash: %v", err)
	}

	// Create our ranges
	ranges := []HashedPeriod{
		// Users (no time range)
		{Hash: usersHash},

		// Current partial week
		{
			Start: &thisWeekStart,
			End:   &now,
			Hash:  currentWeekHash,
		},

		// Previous week
		{
			Start: &lastWeekStart,
			End:   &thisWeekStart,
			Hash:  lastWeekHash,
		},

		// Previous four weeks
		{
			Start: &fourWeeksStart,
			End:   &lastWeekStart,
			Hash:  fourWeeksHash,
		},

		// Previous six months
		{
			Start: &sixMonthsStart,
			End:   &fourWeeksStart,
			Hash:  sixMonthsHash,
		},

		// Everything before six months
		{
			End:  &sixMonthsStart,
			Hash: oldestHash,
		},
	}

	return ranges, nil
}

// SplitTimeRange splits a time range into n equal parts
func SplitTimeRange(start, end time.Time, n int) []HashedPeriod {
	duration := end.Sub(start)
	partDuration := duration / time.Duration(n)

	ranges := make([]HashedPeriod, n)
	for i := 0; i < n; i++ {
		partStart := start.Add(partDuration * time.Duration(i))
		partEnd := partStart.Add(partDuration)
		if i == n-1 {
			partEnd = end // Ensure we don't miss any time due to rounding
		}

		ranges[i] = HashedPeriod{
			Start: &partStart,
			End:   &partEnd,
		}
	}

	return ranges
}
