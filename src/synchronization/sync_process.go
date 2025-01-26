package synchronization

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"time"

	"axial/database"
	"axial/models"
)

func StartSync(node models.RemoteNode) error {
	hash, err := models.GetDatabaseHash(database.DB)
	if err != nil {
		return err
	}

	if hash == node.Hash {
		return nil
	}

	if !models.StartSync() {
		return fmt.Errorf("failed to start sync")
	}
	defer models.EndSync()

	periods := startingSyncRanges()
	hashedPeriods, err := models.GenerateHashRanges(database.DB, periods)
	if err != nil {
		return err
	}

	messages, err := Sync(node, hashedPeriods)
	if err != nil {
		return err
	}

	for _, message := range messages {
		// Send messages unique to this node to the remote node
		err := SendMessage(node, message)
		if err != nil {
			return err
		}
	}

	return nil
}

func Sync(node models.RemoteNode, hashedPeriods []models.HashedPeriod) ([]models.Message, error) {
	
	syncRequest := models.SyncRequest{
		Ranges: hashedPeriods,
	}

	jsonRequest, err := json.Marshal(syncRequest)
	if err != nil {
		return []models.Message{}, err
	}

	response, err := http.Post(fmt.Sprintf("http://%s/v1/sync", node.Address), "application/json", bytes.NewBuffer(jsonRequest))
	if err != nil {
		return []models.Message{}, fmt.Errorf("failed to send sync request: %v", err)
	}
	defer response.Body.Close()

	var syncResponse models.SyncResponse
	err = json.NewDecoder(response.Body).Decode(&syncResponse)
	if err != nil {
		return []models.Message{}, fmt.Errorf("failed to decode sync response: %v", err)
	}

	if syncResponse.IsBusy {
		// Wait until another time.
		return []models.Message{}, nil
	}

	messagesMissingInRemote := []models.Message{}
	
	for _, messagesPeriod := range syncResponse.Messages {
		ourMessages, err := models.GetMessagesByPeriod(database.DB, messagesPeriod.Period)
		if err != nil {
			return []models.Message{}, fmt.Errorf("failed to get messages by period: %v", err)
		}

		for _, message := range messagesPeriod.Messages {
			if !slices.Contains(ourMessages, message) {
				// Insert message into our database
				if err := database.DB.Create(&message).Error; err != nil {
					return []models.Message{}, err
				}
			}
		}

		for _, message := range ourMessages {
			if !slices.Contains(messagesPeriod.Messages, message) {
				messagesMissingInRemote = append(messagesMissingInRemote, message)
			}
		}
	}

	periodsForRemoteHashes := []models.Period{}
	for _, hashedPeriod := range syncResponse.Ranges {
		periodsForRemoteHashes = append(periodsForRemoteHashes, hashedPeriod.Period)
	}

	ourHashes, err := models.GenerateHashRanges(database.DB, periodsForRemoteHashes)
	if err != nil {
		return []models.Message{}, fmt.Errorf("failed to generate hash ranges: %v", err)
	}


	hashedPeriodsToCheck := []models.HashedPeriod{}

	for _, ourHash := range ourHashes {
		start := models.RealizeStart(ourHash.Start)
		end := models.RealizeEnd(ourHash.End)
		for _, theirHash := range syncResponse.Ranges {
			theirStart := models.RealizeStart(theirHash.Start)
			theirEnd := models.RealizeEnd(theirHash.End)
			if theirStart == start && theirEnd == end && theirHash.Hash != ourHash.Hash {
				hashedPeriodsToCheck = append(hashedPeriodsToCheck, theirHash)
			}
		}
	}

	newMessagesMissingInRemote, err := Sync(node, hashedPeriodsToCheck)
	if err != nil {
		return []models.Message{}, fmt.Errorf("failed to sync new messages missing in remote: %v", err)
	}

	messagesMissingInRemote = append(messagesMissingInRemote, newMessagesMissingInRemote...)
	
	return messagesMissingInRemote, nil
}

func startingSyncRanges() []models.Period {

	earliestStartTime := models.RealizeStart(nil)

	periods := []models.Period{}
	// The current week starting on monday
	weekStart := getWeekStart()
	if weekStart.Before(earliestStartTime) {
		return []models.Period{
			{
			},
		}
	}
	periods = append(periods, models.Period{
		Start: &weekStart,
	})

	// The month before the current week
	monthStart := weekStart.AddDate(0, -1, 0)
	if monthStart.Before(earliestStartTime) {
		periods = append(periods, models.Period{
			End: &weekStart,
		})
		return periods
	}

	periods = append(periods, models.Period{
		Start: &monthStart,
		End:   &weekStart,
	})

	// The 6 months before that month
	sixMonths := monthStart.AddDate(0, -6, 0)
	if sixMonths.Before(earliestStartTime) {
		periods = append(periods, models.Period{
			End: &monthStart,
		})
		return periods
	}
	periods = append(periods, models.Period{
		Start: &sixMonths,
		End:   &monthStart,
	})

	// Two years before that six months
	twoYears := sixMonths.AddDate(-2, 0, 0)
	if twoYears.Before(earliestStartTime) {
		periods = append(periods, models.Period{
			End: &sixMonths,
		})
		return periods
	}
	periods = append(periods, models.Period{
		Start: &twoYears,
		End:   &sixMonths,
	})

	// Everything before that
	periods = append(periods, models.Period{
		End:   &twoYears,
	})

	return periods
}


func getWeekStart() time.Time {
	now := time.Now()
	weekday := now.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	return now.AddDate(0, 0, -int(weekday-time.Monday)).Truncate(24 * time.Hour)
}