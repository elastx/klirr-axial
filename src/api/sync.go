package api

import (
	"axial/models"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"gorm.io/gorm"
)

const (
	maxBatchSize   = 1000 // Maximum number of messages to return in one response
	numRangeSplits = 10   // Number of parts to split a range into when too large
)

type SyncRequest struct {
	Ranges []models.HashedPeriod `json:"ranges"`
	Users  []models.StringRange  `json:"users"`
}

type SyncResponse struct {
	Hashes          models.HashSet            `json:"hash"`
	IsBusy          bool                      `json:"is_busy"`
	Ranges          []models.HashedPeriod     `json:"ranges,omitempty"`
	Messages        []models.MessagesPeriod   `json:"messages,omitempty"`
	UserRangeHashes []models.HashedUsersRange `json:"user_range_hashes,omitempty"`
	Users           []models.UsersRange       `json:"users,omitempty"`
}

func handleSync(w http.ResponseWriter, r *http.Request) {
	// Check if we're busy
	if models.IsSyncing() {
		json.NewEncoder(w).Encode(SyncResponse{
			IsBusy: true,
		})
		return
	}

	if r.Method == http.MethodPost {
		handleSyncRequest(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleSyncRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Handling sync request...\n")
	if !models.StartSync() {
		fmt.Printf("Sync already in progress, returning busy response\n")
		json.NewEncoder(w).Encode(SyncResponse{
			IsBusy: true,
		})
		return
	}
	defer models.EndSync()

	var req SyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Printf("Failed to decode request body: %v\n", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	resp, err := ComputeSyncResponse(models.DB, req)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(resp)
}

// ComputeSyncResponse encapsulates the core sync logic, producing a response
// for a given request and database. It is used by the HTTP handler and can be
// reused by tests to simulate in-memory sync exchanges without HTTP.
func ComputeSyncResponse(db *gorm.DB, req SyncRequest) (SyncResponse, error) {
	periods := []models.Period{}
	for _, period := range req.Ranges {
		periods = append(periods, models.Period{
			Start: period.Start,
			End:   period.End,
		})
	}
	fmt.Printf("Received %d time ranges to check\n", len(periods))

	// Generate our hashes for the same ranges
	ourRanges, err := models.GenerateHashRanges(db, periods)
	if err != nil {
		return SyncResponse{}, fmt.Errorf("failed to generate hash ranges: %v", err)
	}
	fmt.Printf("Generated %d hash ranges\n", len(ourRanges))

	mismatchingRanges := []models.HashedPeriod{}
	for _, ourRange := range ourRanges {
		ourStart := models.RealizeStart(ourRange.Start)
		ourEnd := models.RealizeEnd(ourRange.End)
		for _, theirRange := range req.Ranges {
			theirStart := models.RealizeStart(theirRange.Start)
			theirEnd := models.RealizeEnd(theirRange.End)
			if ourStart == theirStart && ourEnd == theirEnd {
				fmt.Printf("Checking period ours: %v to %v, theirs: %v to %v\n", ourStart, ourEnd, theirStart, theirEnd)
				if ourRange.Hash != theirRange.Hash {
					fmt.Printf("Found mismatching hash for range %v to %v (our hash: %s, their hash: %s)\n",
						ourStart, ourEnd, ourRange.Hash, theirRange.Hash)
					mismatchingRanges = append(mismatchingRanges, ourRange)
				} else {
					fmt.Printf("No mismatching hash for range %v to %v (our hash: %s, their hash: %s)\n",
						ourStart, ourEnd, ourRange.Hash, theirRange.Hash)
				}
			} else {
				fmt.Printf("Periods do not match: ours: %v to %v, theirs: %v to %v\n",
					ourStart, ourEnd, theirStart, theirEnd)
			}
		}
	}
	fmt.Printf("Found %d mismatching ranges\n", len(mismatchingRanges))

	// Compare hashes and prepare response
	resp := SyncResponse{}
	resp.Hashes, err = models.GetDatabaseHashes(db)
	if err != nil {
		return SyncResponse{}, fmt.Errorf("failed to get database hash: %v", err)
	}
	fmt.Printf("Our database hashes: %+v\n", resp.Hashes)

	counts := map[int]int64{}
	for index, mismatchingRange := range mismatchingRanges {
		period := models.Period{
			Start: mismatchingRange.Start,
			End:   mismatchingRange.End,
		}
		counts[index] = models.CountMessagesByPeriod(db, period)
		fmt.Printf("Range %d has %d messages\n", index, counts[index])
	}

	// Sort indices by count in ascending order
	indicesSortedByCount := make([]int, 0, len(counts))
	for index := range counts {
		indicesSortedByCount = append(indicesSortedByCount, index)
	}
	sort.Slice(indicesSortedByCount, func(i, j int) bool {
		return counts[indicesSortedByCount[i]] < counts[indicesSortedByCount[j]]
	})

	totalPlainMessages := int64(0)

	for _, index := range indicesSortedByCount {
		mismatchingRange := mismatchingRanges[index]
		// Try to return as many messages as possible
		if totalPlainMessages+counts[index] < maxBatchSize {
			fmt.Printf("Getting messages for range %d (count: %d, total so far: %d)\n",
				index, counts[index], totalPlainMessages)
			messages, err := models.GetMessagesByPeriod(db, mismatchingRange.Period)
			if err != nil {
				return SyncResponse{}, fmt.Errorf("failed to get messages: %v", err)
			}

			messagesPeriod := models.MessagesPeriod{
				Period:   mismatchingRange.Period,
				Messages: messages,
			}
			resp.Messages = append(resp.Messages, messagesPeriod)
			totalPlainMessages += counts[index]
		} else {
			fmt.Printf("Range %d too large (%d messages), splitting into smaller ranges\n",
				index, counts[index])
			// All batches that don't fit the plain message limit are returned as more granular
			// hashed ranges, for drilling down to find the mismatching data.
			//
			// The following ensures that each split is 1/10 of the size possible to return
			// in the next step. This is to ensure we get to return actual messages
			// in the next step instead of further hashed range juggling.
			splits := int(counts[index]/(maxBatchSize*10)) + 1
			fmt.Printf("Splitting range into %d parts\n", splits)
			periods := models.SplitTimeRange(mismatchingRange.Period, splits)
			for _, period := range periods {
				hashedPeriods, err := models.GenerateHashRanges(db, []models.Period{period})
				if err != nil {
					return SyncResponse{}, fmt.Errorf("failed to generate hash ranges for split: %v", err)
				}
				for _, hashedPeriod := range hashedPeriods {
					resp.Ranges = append(resp.Ranges, models.HashedPeriod{
						Period: hashedPeriod.Period,
						Hash:   hashedPeriod.Hash,
					})
				}
			}
		}
	}

	return resp, nil
}
