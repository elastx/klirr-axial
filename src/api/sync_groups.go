package api

import (
	"axial/models"
	"encoding/json"
	"net/http"
)

type SyncGroupsRequest struct {
	Groups []models.Group `json:"groups"`
}

func handleSyncGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SyncGroupsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if len(req.Groups) == 0 {
		http.Error(w, "Groups are required", http.StatusBadRequest)
		return
	}

	// Create groups
	for _, group := range req.Groups {
		if err := models.DB.Create(&group).Error; err != nil {
			// Ignore duplicate errors
			if models.IsDuplicateError(err) {
				continue
			}
			http.Error(w, "Failed to create group", http.StatusInternalServerError)
			return
		}
	}

	models.RefreshHashes(models.DB)

	w.WriteHeader(http.StatusCreated)

}