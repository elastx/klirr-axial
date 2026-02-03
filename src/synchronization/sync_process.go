package synchronization

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"axial/api"
	"axial/models"
	"axial/remote"
)

// SyncRequester abstracts how a sync request is sent to a remote node.
// Production uses HTTP; tests can provide an in-memory implementation to
// simulate back-and-forth exchanges without network or servers.
type SyncRequester interface {
	RequestSync(node remote.API, req api.SyncRequest) (api.SyncResponse, error)
}

// httpSyncRequester implements SyncRequester over HTTP to the node's address.
type httpSyncRequester struct {
	Client *http.Client
}

func (h httpSyncRequester) RequestSync(node remote.API, req api.SyncRequest) (api.SyncResponse, error) {
	jsonRequest, err := json.Marshal(req)
	if err != nil {
		return api.SyncResponse{}, err
	}

	response, err := h.httpPost(node, jsonRequest)
	if err != nil {
		return api.SyncResponse{}, fmt.Errorf("failed to send sync request: %v", err)
	}
	defer response.Body.Close()

	var syncResponse api.SyncResponse
	err = json.NewDecoder(response.Body).Decode(&syncResponse)
	if err != nil {
		return api.SyncResponse{}, fmt.Errorf("failed to decode sync response: %v", err)
	}
	return syncResponse, nil
}

func (h httpSyncRequester) httpPost(node remote.API, jsonRequest []byte) (*http.Response, error) {
	client := h.Client
	if client == nil {
		client = http.DefaultClient
	}
	return client.Post(fmt.Sprintf("http://%s/v1/sync", node.Address), "application/json", bytes.NewBuffer(jsonRequest))
}

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

	periods, stringRanges := startingSyncRanges()
	hashedPeriods, err := models.GenerateHashRanges(models.DB, periods)
	if err != nil {
		return err
	}

	hashedUsers, err := models.GetUsersHashRanges(models.DB, stringRanges)
	if err != nil {
		return err
	}

	fmt.Printf("Synchronizing with %s\n", node.Address)

	// Use HTTP requester by default in production flows.
	messages, err := Sync(node, hashedPeriods, hashedUsers)
	if err != nil {
		return err
	}

	// Get the usersMap that sent the messages
	usersMap := map[string]*models.User{}
	for _, message := range messages {
		user, err := models.GetUserByFingerprint(message.Sender)
		if err != nil {
			return err
		}
		usersMap[string(user.Fingerprint)] = user
	}
	if len(usersMap) == 0 {
		usersList := []models.User{}
		for _, user := range usersMap {
			usersList = append(usersList, *user)
		}

		SyncUsers(node, usersList)

		// Sort messages by creation time
		SortMessages(messages)

		// Send messages unique to this node to the remote node
		SyncMessages(node, messages)
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

// Sync performs one round of synchronization with a remote node using the provided
// hash ranges. It returns messages present locally but missing in the remote.
//
// For unit tests, prefer calling SyncWithRequester with a custom requester that
// uses in-memory handlers to return api.SyncResponse.
func Sync(node remote.API, hashedPeriods []models.HashedPeriod, hashedUsers []models.HashedUsersRange) ([]models.Message, error) {
	return SyncWithRequester(httpSyncRequester{}, node, hashedPeriods, hashedUsers)
}

// SyncWithRequester is identical to Sync but allows the caller to provide a
// pluggable requester for testability.
func SyncWithRequester(requester SyncRequester, node remote.API, hashedPeriods []models.HashedPeriod, hashedUsers []models.HashedUsersRange) ([]models.Message, error) {
	if len(hashedPeriods) == 0 {
		fmt.Printf("No periods to sync with %s\n", node.Address)
		return []models.Message{}, nil
	}

	syncRequest := api.SyncRequest{
		Ranges: hashedPeriods,
	}

	// Let the requester handle the transport (HTTP in prod, in-memory in tests).
	fmt.Printf("Sending sync request to %s\n", node.Address)
	syncResponse, err := requester.RequestSync(node, syncRequest)
	if err != nil {
		return []models.Message{}, err
	}

	fmt.Printf("Received sync response from %s: %+v\n", node.Address, syncResponse)

	if syncResponse.IsBusy {
		// Wait until another time.
		return []models.Message{}, nil
	}

	messagesMissingInRemote := []models.Message{}

	for _, messagesPeriod := range syncResponse.Messages {
		ourMessages, err := models.GetMessagesByPeriod(models.DB, messagesPeriod.Period)
		if err != nil {
			return []models.Message{}, fmt.Errorf("failed to get messages by period: %v", err)
		}

		for _, message := range messagesPeriod.Messages {
			if !message.In(ourMessages) {
				fmt.Printf("Inserting message into our database: %+v\n", message)
				// Insert message into our database
				if err := models.DB.Create(&message).Error; err != nil {
					// Ignore duplicate key errors since those messages were already synced
					if !strings.Contains(err.Error(), "duplicate key") {
						return []models.Message{}, err
					}
				}
			}
		}

		for _, message := range ourMessages {
			if !message.In(messagesPeriod.Messages) {
				messagesMissingInRemote = append(messagesMissingInRemote, message)
			}
		}
	}

	periodsForRemoteHashes := []models.Period{}
	for _, hashedPeriod := range syncResponse.Ranges {
		periodsForRemoteHashes = append(periodsForRemoteHashes, hashedPeriod.Period)
	}

	ourHashes, err := models.GenerateHashRanges(models.DB, periodsForRemoteHashes)
	if err != nil {
		return []models.Message{}, fmt.Errorf("failed to generate hash ranges: %v", err)
	}

	hashedPeriodsToCheck := mismatchedPeriods(ourHashes, syncResponse.Ranges)

	userRangesToCheck := []models.HashedUsersRange{}

	for _, hashedUserRange := range syncResponse.UserRangeHashes {
		ourUserHash, err := models.GetUsersHashByFingerprintRange(models.DB, hashedUserRange.Start, hashedUserRange.End)
		if err != nil {
			return []models.Message{}, fmt.Errorf("failed to get users by fingerprint range: %v", err)
		}

		if ourUserHash != hashedUserRange.Hash {
			userRangesToCheck = append(userRangesToCheck, hashedUserRange)
		}

	}

	newMessagesMissingInRemote, err := SyncWithRequester(requester, node, hashedPeriodsToCheck, userRangesToCheck)
	if err != nil {
		return []models.Message{}, fmt.Errorf("failed to sync new messages missing in remote: %v", err)
	}

	messagesMissingInRemote = append(messagesMissingInRemote, newMessagesMissingInRemote...)

	return messagesMissingInRemote, nil
}

// mismatchedPeriods returns the set of hashed periods from the remote that
// correspond to the same concrete time ranges as ours but have different
// content hashes.
func mismatchedPeriods(our []models.HashedPeriod, theirs []models.HashedPeriod) []models.HashedPeriod {
	out := []models.HashedPeriod{}
	for _, ourHash := range our {
		start := models.RealizeStart(ourHash.Start)
		end := models.RealizeEnd(ourHash.End)
		for _, theirHash := range theirs {
			theirStart := models.RealizeStart(theirHash.Start)
			theirEnd := models.RealizeEnd(theirHash.End)
			if theirStart == start && theirEnd == end && theirHash.Hash != ourHash.Hash {
				out = append(out, theirHash)
			}
		}
	}
	return out
}

func startingSyncRanges() ([]models.Period, []models.StringRange) {

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
	weekStart := getWeekStart()
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
			Start: fmt.Sprintf("%c", '0'+i),
			End:   fmt.Sprintf("%c", '0'+i+1),
		})
	}
	for i := 0; i < 25; i++ {
		userRanges = append(userRanges, models.StringRange{
			Start: fmt.Sprintf("%c", 'a'+i),
			End:   fmt.Sprintf("%c", 'a'+i+1),
		})
	}
	return periods, userRanges
}

func getWeekStart() time.Time {
	now := time.Now()
	weekday := now.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	return now.AddDate(0, 0, -int(weekday-time.Monday)).Truncate(24 * time.Hour)
}
