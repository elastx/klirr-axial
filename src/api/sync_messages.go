package api

import (
	"encoding/json"
	"net/http"

	"axial/database"
	"axial/models"
)

type SyncMessagesRequest struct {
	Messages []models.Message `json:"messages"`
}

func handleSyncMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SyncMessagesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if len(req.Messages) == 0 {
		http.Error(w, "Messages are required", http.StatusBadRequest)
		return
	}

	// Create messages
	for _, message := range req.Messages {
		if err := database.DB.Create(&message).Error; err != nil {
			// Ignore duplicate errors
			if database.IsDuplicateError(err) {
				continue
			}
			http.Error(w, "Failed to create message", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)

}