package api

import (
	"encoding/json"
	"net/http"
	"sort"

	"axial/database"
	"axial/models"
)

const (
	maxBatchSize = 1000 // Maximum number of messages to return in one response
	numRangeSplits = 10 // Number of parts to split a range into when too large
)

func handleSync(w http.ResponseWriter, r *http.Request) {
	// Check if we're busy
	if models.IsSyncing() {
		json.NewEncoder(w).Encode(models.SyncResponse{
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
	if !models.StartSync() {
		json.NewEncoder(w).Encode(models.SyncResponse{
			IsBusy: true,
		})
		return
	}
	defer models.EndSync()

	var req models.SyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	periods := []models.Period{}
	for _, period := range req.Ranges {
		periods = append(periods, models.Period{
			Start: period.Start,
			End:   period.End,
		})
	}

	// Generate our hashes for the same ranges
	ourRanges, err := models.GenerateHashRanges(database.DB, periods)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	mismatchingRanges := []models.HashedPeriod{}
	for _, ourRange := range ourRanges {
		ourStart := models.RealizeStart(ourRange.Start)
		ourEnd := models.RealizeEnd(ourRange.End)
		for _, theirRange := range req.Ranges {
			theirStart := models.RealizeStart(theirRange.Start)
			theirEnd := models.RealizeEnd(theirRange.End)
			if ourStart == theirStart && ourEnd == theirEnd && ourRange.Hash != theirRange.Hash {
				mismatchingRanges = append(mismatchingRanges, ourRange)
			}
		}
	}

	// Compare hashes and prepare response
	var resp models.SyncResponse
	resp.Hash, err = models.GetDatabaseHash(database.DB)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	counts := map[int]int64{}
	for index, mismatchingRange := range mismatchingRanges {
		period := models.Period{
			Start: mismatchingRange.Start,
			End:   mismatchingRange.End,
		}
		counts[index] = models.CountMessagesByPeriod(database.DB, period)
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
		if totalPlainMessages + counts[index] < maxBatchSize {
			messages, err := models.GetMessagesByPeriod(database.DB, mismatchingRange.Period)
			if err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}

			messagesPeriod := models.MessagesPeriod{
				Period: mismatchingRange.Period,
				Messages: messages,
			}
			resp.Messages = append(resp.Messages, messagesPeriod)
			totalPlainMessages += counts[index]
		} else {
			// All batches that don't fit the plain message limit are returned as more granular
			// hashed ranges, for drilling down to find the mismatching data.
			// 
			// The following ensures that each split is 1/10 of the size possible to return
			// in the next step. This is to ensure we get to return actual messages
			// in the next step instead of further hashed range juggling.
			splits := int(counts[index]/(maxBatchSize * 10)) + 1
			periods := models.SplitTimeRange(mismatchingRange.Period, splits)
			for _, period := range periods {
				hashedPeriods, err := models.GenerateHashRanges(database.DB, []models.Period{period})
				if err != nil {
					http.Error(w, "Database error", http.StatusInternalServerError)
					return
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

	json.NewEncoder(w).Encode(resp)
}

