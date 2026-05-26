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
	hashedMessagesPeriods, err := models.GetMessagesHashRanges(models.DB, periods)
	if err != nil {
		return err
	}

	hashedBulletinsPeriods, err := models.GetBulletinsHashRanges(models.DB, periods)
	if err != nil {
		return err
	}

	hashedUsers, err := models.GetUsersHashRanges(models.DB, stringRanges)
	if err != nil {
		return err
	}

	fmt.Printf("Synchronizing with %s\n", node.Address)

	// Use HTTP requester by default in production flows.
	messages, bulletins, users, err := Sync(node, hashedMessagesPeriods, hashedBulletinsPeriods, hashedUsers)
	if err != nil {
		return err
	}

	pushDeltas(node, users, messages, bulletins)
	return nil
}

// pushDeltas POSTs items the local Node has that the remote Node lacks,
// one HTTP call per resource type. Errors are logged but not propagated —
// sync rounds are eventually-consistent; a transient push failure recovers
// on the next round. Messages and bulletins are sorted by CreatedAt before
// the push so the remote ingests in deterministic order.
func pushDeltas(node remote.API, users []models.User, messages []models.Message, bulletins []models.Bulletin) {
	if len(users) > 0 {
		ep := node.SyncUsers()
		_, resp, err := ep.Post(users)
		logPush(node, "users", len(users), resp, err)
	}

	SortMessages(messages)
	if len(messages) > 0 {
		ep := node.SyncMessages()
		_, resp, err := ep.Post(messages)
		logPush(node, "messages", len(messages), resp, err)
	}

	SortBulletins(bulletins)
	if len(bulletins) > 0 {
		ep := node.SyncBulletins()
		_, resp, err := ep.Post(bulletins)
		logPush(node, "bulletins", len(bulletins), resp, err)
	}
}

func logPush(node remote.API, kind string, count int, resp *http.Response, err error) {
	if err != nil {
		fmt.Printf("Failed to push %d %s to %s: %v\n", count, kind, node.Address, err)
		return
	}
	if resp != nil {
		fmt.Printf("Pushed %d %s to %s: %s\n", count, kind, node.Address, resp.Status)
	}
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

func SortBulletins(bulletins []models.Bulletin) {
	for i := 0; i < len(bulletins); i++ {
		for j := i + 1; j < len(bulletins); j++ {
			if bulletins[i].CreatedAt.After(bulletins[j].CreatedAt) {
				bulletins[i], bulletins[j] = bulletins[j], bulletins[i]
			}
		}
	}
}

// Sync performs one round of synchronization with a remote node using the provided
// hash ranges. It returns messages present locally but missing in the remote.
//
// For unit tests, prefer calling SyncWithRequester with a custom requester that
// uses in-memory handlers to return api.SyncResponse.
func Sync(node remote.API, hashedMessagePeriods []models.HashedPeriod, hashedBulletinPeriods []models.HashedPeriod, hashedUsers []models.HashedUsersRange) ([]models.Message, []models.Bulletin, []models.User, error) {
	return SyncWithRequester(httpSyncRequester{}, node, hashedMessagePeriods, hashedBulletinPeriods, hashedUsers)
}

// SyncWithRequester is identical to Sync but allows the caller to provide a
// pluggable requester for testability.
func SyncWithRequester(requester SyncRequester, node remote.API, hashedMessagesPeriods []models.HashedPeriod, hashedBulletinPeriods []models.HashedPeriod, hashedUsers []models.HashedUsersRange) ([]models.Message, []models.Bulletin, []models.User, error) {
	if len(hashedMessagesPeriods) == 0 {
		fmt.Printf("No periods to sync with %s\n", node.Address)
		return []models.Message{}, []models.Bulletin{}, []models.User{}, nil
	}

	syncRequest := api.SyncRequest{
		MessageRanges:  hashedMessagesPeriods,
		BulletinRanges: hashedBulletinPeriods,
		Users:          hashedUsers,
	}

	// Let the requester handle the transport (HTTP in prod, in-memory in tests).
	fmt.Printf("Sending sync request to %s\n", node.Address)
	syncResponse, err := requester.RequestSync(node, syncRequest)
	if err != nil {
		return []models.Message{}, []models.Bulletin{}, []models.User{}, err
	}

	fmt.Printf("Received sync response from %s: %+v\n", node.Address, syncResponse)

	if syncResponse.IsBusy {
		// Wait until another time.
		return []models.Message{}, []models.Bulletin{}, []models.User{}, nil
	}

	// Messages
	messagesMissingInRemote := []models.Message{}

	for _, messagesPeriod := range syncResponse.Messages {
		period := messagesPeriod.Period
		missing, err := reconcile[models.Message](
			func() ([]models.Message, error) {
				return models.GetMessagesByPeriod(models.DB, period)
			},
			messagesPeriod.Messages,
			func(a, b models.Message) bool { return a.ID == b.ID },
			func(m models.Message) error { return models.DB.Create(&m).Error },
			isDuplicateKeyErr,
		)
		if err != nil {
			return []models.Message{}, []models.Bulletin{}, []models.User{}, fmt.Errorf("reconcile messages: %v", err)
		}
		messagesMissingInRemote = append(messagesMissingInRemote, missing...)
	}

	periodsForRemoteMessagesHashes := []models.Period{}
	for _, hashedPeriod := range syncResponse.MessageRanges {
		periodsForRemoteMessagesHashes = append(periodsForRemoteMessagesHashes, hashedPeriod.Period)
	}

	ourMessagesHashes, err := models.GetMessagesHashRanges(models.DB, periodsForRemoteMessagesHashes)
	if err != nil {
		return []models.Message{}, []models.Bulletin{}, []models.User{}, fmt.Errorf("failed to generate hash ranges: %v", err)
	}

	hashedMessagesPeriodsToCheck := mismatchedMessagesPeriods(ourMessagesHashes, syncResponse.MessageRanges)

	// Bulletins
	bulletinsMissingInRemote := []models.Bulletin{}

	for _, bulletinPeriod := range syncResponse.Bulletins {
		period := bulletinPeriod.Period
		missing, err := reconcile[models.Bulletin](
			func() ([]models.Bulletin, error) {
				return models.GetBulletinsByPeriod(models.DB, period)
			},
			bulletinPeriod.Bulletins,
			func(a, b models.Bulletin) bool { return a.ID == b.ID },
			func(b models.Bulletin) error { return models.DB.Create(&b).Error },
			isDuplicateKeyErr,
		)
		if err != nil {
			return []models.Message{}, []models.Bulletin{}, []models.User{}, fmt.Errorf("reconcile bulletins: %v", err)
		}
		bulletinsMissingInRemote = append(bulletinsMissingInRemote, missing...)
	}

	periodsForRemoteBulletinHashes := []models.Period{}
	for _, hashedPeriod := range syncResponse.BulletinRanges {
		periodsForRemoteBulletinHashes = append(periodsForRemoteBulletinHashes, hashedPeriod.Period)
	}

	ourBulletinHashes, err := models.GetBulletinsHashRanges(models.DB, periodsForRemoteBulletinHashes)
	if err != nil {
		return []models.Message{}, []models.Bulletin{}, []models.User{}, fmt.Errorf("failed to generate bulletin hash ranges: %v", err)
	}

	hashedBulletinPeriodsToCheck := mismatchedMessagesPeriods(ourBulletinHashes, syncResponse.BulletinRanges)

	// Users — equality is ID-or-fingerprint-group; sameUserGroup is needed
	// because the integration tests synthesise fingerprints like
	// "FP_A2_<random>" and expect peers to recognise the shared identity.
	usersMissingInRemote := []models.User{}
	userEqual := func(a, b models.User) bool {
		return a.ID == b.ID || sameUserGroup(a.Fingerprint, b.Fingerprint)
	}

	for _, usersRange := range syncResponse.Users {
		ur := usersRange
		missing, err := reconcile[models.User](
			func() ([]models.User, error) {
				return models.GetUsersByFingerprintRange(models.DB, ur.StringRange.Start, ur.StringRange.End)
			},
			ur.Users,
			userEqual,
			func(u models.User) error { return models.DB.Create(&u).Error },
			isDuplicateKeyErr,
		)
		if err != nil {
			return []models.Message{}, []models.Bulletin{}, []models.User{}, fmt.Errorf("reconcile users: %v", err)
		}
		usersMissingInRemote = append(usersMissingInRemote, missing...)
	}

	userRangesToCheck := []models.HashedUsersRange{}

	for _, hashedUserRange := range syncResponse.UserRangeHashes {
		ourUserHash, err := models.GetUsersHashByFingerprintRange(models.DB, hashedUserRange.Start, hashedUserRange.End)
		if err != nil {
			return []models.Message{}, []models.Bulletin{}, []models.User{}, fmt.Errorf("failed to get users by fingerprint range: %v", err)
		}

		if ourUserHash != hashedUserRange.Hash {
			userRangesToCheck = append(userRangesToCheck, hashedUserRange)
		}

	}

	newMessagesMissingInRemote, newBulletinsMissingInRemote, newUsersMissingInRemote, err := SyncWithRequester(requester, node, hashedMessagesPeriodsToCheck, hashedBulletinPeriodsToCheck, userRangesToCheck)
	if err != nil {
		return []models.Message{}, []models.Bulletin{}, []models.User{}, fmt.Errorf("failed to sync new messages missing in remote: %v", err)
	}

	messagesMissingInRemote = append(messagesMissingInRemote, newMessagesMissingInRemote...)
	bulletinsMissingInRemote = append(bulletinsMissingInRemote, newBulletinsMissingInRemote...)
	usersMissingInRemote = append(usersMissingInRemote, newUsersMissingInRemote...)

	return messagesMissingInRemote, bulletinsMissingInRemote, usersMissingInRemote, nil
}

// sameUserGroup attempts to determine if two fingerprints represent the same
// logical user by comparing the prefix up to the second underscore. This helps
// tests that generate synthetic fingerprints like "FP_A2_<random>" to avoid
// duplicating logically shared identities across peers.
func sameUserGroup(a string, b string) bool {
	ap := userGroupPrefix(a)
	bp := userGroupPrefix(b)
	return ap != "" && ap == bp
}

func userGroupPrefix(s string) string {
	// Find up to second underscore
	first := strings.Index(s, "_")
	if first < 0 {
		return ""
	}
	second := strings.Index(s[first+1:], "_")
	if second < 0 {
		return s[:first]
	}
	// include part before second underscore
	return s[:first+1+second]
}

// mismatchedMessagesPeriods returns the set of hashed periods from the remote that
// correspond to the same concrete time ranges as ours but have different
// content hashes.
func mismatchedMessagesPeriods(our []models.HashedPeriod, theirs []models.HashedPeriod) []models.HashedPeriod {
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
