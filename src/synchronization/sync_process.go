package synchronization

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"axial/api"
	"axial/models"
	"axial/remote"
)

func StartSync(node remote.API, hash string) error {
	hashes, err := models.GetDatabaseHashes(models.DB)
	if err != nil {
		return err
	}

	if hashes.Full == hash {
		return nil
	}

	if !models.StartSync() {
		return fmt.Errorf("failed to start sync")
	}
	defer models.EndSync()

	periods, stringRanges := startingSyncRanges(DefaultClock)
	hashedPeriods, err := models.GenerateHashRanges(models.DB, periods)
	if err != nil {
		return err
	}

	hashedUsers, err := models.GetUsersHashRanges(models.DB, stringRanges)
	if err != nil {
		return err
	}

	fmt.Printf("Synchronizing with %s\n", node.Address)

	messages, err := Sync(node, hashedPeriods, hashedUsers)
	if err != nil {
		return err
	}

	// Sync all users, not just senders
	allUsers := []models.User{}
	if err := models.DB.Find(&allUsers).Error; err != nil {
		return err
	}
	if len(allUsers) > 0 {
		if err := SyncUsers(node, allUsers); err != nil {
			return err
		}
	}

	// Sort and sync messages unique to this node to the remote
	if len(messages) > 0 {
		SortMessages(messages)
		if err := SyncMessages(node, messages); err != nil {
			return err
		}
	}

	// Sync bulletin board items
	bulletins := []models.Bulletin{}
	if err := models.DB.Find(&bulletins).Error; err != nil {
		return err
	}
	if len(bulletins) > 0 {
		if err := SyncBulletin(node, bulletins); err != nil {
			return err
		}
	}

	return nil
}

func SortMessages(messages []models.Message) {
	for i := 0; i < len(messages); i++ {
		for j := i + 1; j < len(messages); j++ {
			if messages[i].CreatedAt.After(messages[j].CreatedAt) {
				messages[i], messages[j] = messages[j], messages[i]
			}
		}
	}
}

func Sync(node remote.API, hashedPeriods []models.HashedPeriod, hashedUsers []models.HashedUsersRange) ([]models.Message, error) {
	if len(hashedPeriods) == 0 {
		fmt.Printf("No periods to sync with %s\n", node.Address)
		return []models.Message{}, nil
	}

	store := &api.ModelSyncStore{}
	messagesMissingInRemote := []models.Message{}

	// Iterative drill-down loop
	nextPeriods := hashedPeriods
	nextUsers := hashedUsers

	for {
		syncRequest := api.SyncRequest{Ranges: nextPeriods, Users: nextUsers}
		jsonRequest, err := json.Marshal(syncRequest)
		if err != nil {
			return []models.Message{}, err
		}

		fmt.Printf("Sending sync request to %s: %s\n", node.Address, string(jsonRequest))

		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/v1/sync", node.Address), bytes.NewBuffer(jsonRequest))
		if err != nil {
			return []models.Message{}, fmt.Errorf("failed to build sync request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		response, err := Client.Do(req)
		if err != nil {
			return []models.Message{}, fmt.Errorf("failed to send sync request: %v", err)
		}
		var syncResponse api.SyncResponse
		err = json.NewDecoder(response.Body).Decode(&syncResponse)
		response.Body.Close()
		if err != nil {
			return []models.Message{}, fmt.Errorf("failed to decode sync response: %v", err)
		}

		if syncResponse.IsBusy {
			return messagesMissingInRemote, nil
		}

		// Apply incoming messages
		if err := ApplyIncomingMessages(store, syncResponse.Messages); err != nil {
			return []models.Message{}, err
		}

		// Accumulate messages missing on remote
		localMissing, err := CollectMissingInRemote(store, syncResponse.Messages)
		if err != nil {
			return []models.Message{}, err
		}
		messagesMissingInRemote = append(messagesMissingInRemote, localMissing...)

		// Determine next drill-down request
		periodsToCheck, usersToCheck, err := NextRequestFromHashes(store, syncResponse)
		if err != nil {
			return []models.Message{}, err
		}

		if len(periodsToCheck) == 0 && len(usersToCheck) == 0 {
			break
		}
		nextPeriods = periodsToCheck
		nextUsers = usersToCheck
	}

	return messagesMissingInRemote, nil
}

func startingSyncRanges(clock Clock) ([]models.Period, []models.StringRange) {
	earliestStartTime := models.RealizeStart(nil)
	latestEndTime := models.RealizeEnd(nil)

	periodSteps := []struct {
		Years  int `json:"years"`
		Months int `json:"months"`
		Days   int `json:"days"`
	}{
		{0, -1, 0},
		{0, -6, 0},
		{-2, 0, 0},
	}
	var previousStart *time.Time
	weekStart := getWeekStart(clock)
	if weekStart.Before(earliestStartTime) {
		previousStart = &earliestStartTime
	} else {
		previousStart = &weekStart
	}

	periods := []models.Period{
		{
			Start: previousStart,
			End:   &latestEndTime,
		},
	}

	for _, step := range periodSteps {
		start := previousStart.AddDate(step.Years, step.Months, step.Days)
		if start.Before(earliestStartTime) {
			periods = append(periods, models.Period{
				Start: &earliestStartTime,
				End:   previousStart,
			})
			break
		}
		periods = append(periods, models.Period{
			Start: &start,
			End:   previousStart,
		})
		previousStart = &start
	}

	periods = append(periods, models.Period{
		Start: &earliestStartTime,
		End:   previousStart,
	})

	// Generate user fingerprint ranges, an array of 0-9 and a-z
	var userRanges []models.StringRange
	for i := 0; i < 10; i++ {
		userRanges = append(userRanges, models.StringRange{
			Start: string(rune('0' + i)),
			End:   string(rune('0' + i + 1)),
		})
	}
	for i := 0; i < 25; i++ {
		userRanges = append(userRanges, models.StringRange{
			Start: string(rune('a' + i)),
			End:   string(rune('a' + i + 1)),
		})
	}
	return periods, userRanges
}

func getWeekStart(clock Clock) time.Time {
	now := clock.Now()
	weekday := now.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	return now.AddDate(0, 0, -int(weekday-time.Monday)).Truncate(24 * time.Hour)
}
