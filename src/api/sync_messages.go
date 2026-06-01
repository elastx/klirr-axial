package api

import (
	"encoding/json"
	"net/http"

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

	if len(req.Messages) == 0 {
		http.Error(w, "Messages are required", http.StatusBadRequest)
		return
	}

	if err := ingest(models.DB, req.Messages); err != nil {
		http.Error(w, "Failed to ingest messages", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
