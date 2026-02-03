package api

import (
	"axial/models"
	"encoding/json"
	"net/http"
)

type SyncUsersRequest struct {
	Users []models.User `json:"users"`
}

func handleSyncUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SyncUsersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if len(req.Users) == 0 {
		http.Error(w, "Users are required", http.StatusBadRequest)
		return
	}

	// Create users
	for _, user := range req.Users {
		if err := models.DB.Create(&user).Error; err != nil {
			// Ignore duplicate errors
			if models.IsDuplicateError(err) {
				continue
			}
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
	}

	models.RefreshHashes(models.DB)

	w.WriteHeader(http.StatusCreated)

}