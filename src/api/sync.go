package api

import (
	"axial/models"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	maxBatchSize   = 1000 // Maximum number of messages to return in one response
	numRangeSplits = 10   // Number of parts to split a range into when too large
)

type SyncRequest struct {
	Ranges []models.HashedPeriod `json:"ranges"`
	Users  []models.HashedUsersRange  `json:"users"`
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
	// Delegate to the pure engine using the default DB-backed store
	store := &ModelSyncStore{}
	resp, err := BuildSyncResponse(store, req, maxBatchSize)
	if err != nil {
		fmt.Printf("Failed to build sync response: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(resp)
}
