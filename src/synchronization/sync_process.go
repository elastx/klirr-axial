package synchronization

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
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
	
	fmt.Printf("Synchronizing with %s\n", node.Address)

	messages, err := Sync(node, hashedPeriods)
	if err != nil {
		return err
	}

	// Get the usersMap that sent the messages
	usersMap := map[string]*models.User{}
	for _, message := range messages {
		user, err := database.GetUserByFingerprint(message.Author)
		if err != nil {
			return err
		}
		usersMap[user.Fingerprint] = user
	}

	usersList := []models.User{}
	for _, user := range usersMap {
		usersList = append(usersList, *user)
	}

	SyncUsers(node, usersList)

	// Sort messages by creation time
	SortMessages(messages)

	// Send messages unique to this node to the remote node
	SyncMessages(node, messages)

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

func Sync(node models.RemoteNode, hashedPeriods []models.HashedPeriod) ([]models.Message, error) {
	if len(hashedPeriods) == 0 {
		fmt.Printf("No periods to sync with %s\n", node.Address)
		return []models.Message{}, nil
	}

	
	syncRequest := models.SyncRequest{
		Ranges: hashedPeriods,
	}

	jsonRequest, err := json.Marshal(syncRequest)
	if err != nil {
		return []models.Message{}, err
	}

	fmt.Printf("Sending sync request to %s: %s\n", node.Address, string(jsonRequest))

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

	fmt.Printf("Received sync response from %s: %+v\n", node.Address, syncResponse)

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
				fmt.Printf("Inserting message into our database: %+v\n", message)
				// Insert message into our database
				if err := database.DB.Create(&message).Error; err != nil {
					// Ignore duplicate key errors since those messages were already synced
					if !strings.Contains(err.Error(), "duplicate key") {
						return []models.Message{}, err
					}
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
	latestEndTime := models.RealizeEnd(nil)

	periods := []models.Period{}
	// The current week starting on monday
	weekStart := getWeekStart()
	if weekStart.Before(earliestStartTime) {
		return []models.Period{
			{
				Start: &earliestStartTime,
				End: &latestEndTime,
			},
		}
	}
	periods = append(periods, models.Period{
		Start: &weekStart,
		End: &latestEndTime,
	})

	// The month before the current week
	monthStart := weekStart.AddDate(0, -1, 0)
	if monthStart.Before(earliestStartTime) {
		periods = append(periods, models.Period{
			Start: &earliestStartTime,
			End: &monthStart,
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
			Start: &earliestStartTime,
			End: &sixMonths,
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
			Start: &earliestStartTime,
			End: &twoYears,
		})
		return periods
	}
	periods = append(periods, models.Period{
		Start: &twoYears,
		End:   &sixMonths,
	})

	// Everything before that
	periods = append(periods, models.Period{
		Start: &earliestStartTime,
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