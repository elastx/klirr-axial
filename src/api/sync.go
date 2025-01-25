package api

import (
	"encoding/json"
	"net/http"

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
	} else if r.Method == http.MethodPut {
		handleMessageUpdate(w, r)
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

	// Generate our hashes for the same ranges
	ourRanges, err := models.GenerateHashRanges(database.DB)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Compare hashes and prepare response
	var resp models.SyncResponse
	resp.NodeID = req.NodeID
	resp.Hash = models.GetHash()

	// Check each range
	for i, theirRange := range req.Ranges {
		if i >= len(ourRanges) {
			break
		}
		ourRange := ourRanges[i]

		if theirRange.Hash != ourRange.Hash {
			// Get data for this range
			var messages []models.Message
			var users []models.User

			query := database.DB
			if ourRange.Start != nil {
				query = query.Where("created_at >= ?", ourRange.Start)
			}
			if ourRange.End != nil {
				query = query.Where("created_at <= ?", ourRange.End)
			}

			// If this is the users range (no timestamps)
			if ourRange.Start == nil && ourRange.End == nil {
				if err := database.DB.Find(&users).Error; err != nil {
					http.Error(w, "Database error", http.StatusInternalServerError)
					return
				}
				resp.Users = users
				continue
			}

			// Count messages in range
			var count int64
			if err := query.Model(&models.Message{}).Count(&count).Error; err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}

			// If too many messages, split the range
			if count > maxBatchSize {
				splitRanges := models.SplitTimeRange(*ourRange.Start, *ourRange.End, numRangeSplits)
				for i := range splitRanges {
					hash, err := models.GetMessagesHash(database.DB, splitRanges[i].Start, splitRanges[i].End)
					if err != nil {
						http.Error(w, "Database error", http.StatusInternalServerError)
						return
					}
					splitRanges[i].Hash = hash
				}
				resp.Ranges = append(resp.Ranges, splitRanges...)
				continue
			}

			// Get messages in range
			if err := query.Find(&messages).Error; err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
			resp.Messages = append(resp.Messages, messages...)
		}
	}

	json.NewEncoder(w).Encode(resp)
}

func handleMessageUpdate(w http.ResponseWriter, r *http.Request) {
	var msg models.Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Update or create the message
	if err := database.DB.Save(&msg).Error; err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Recalculate database hash
	hash, err := models.GetDatabaseHash(database.DB)
	if err != nil {
		http.Error(w, "Failed to update hash", http.StatusInternalServerError)
		return
	}
	models.UpdateHash(hash)

	w.WriteHeader(http.StatusOK)
} 