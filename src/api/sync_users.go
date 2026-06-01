package api

import (
	"encoding/json"
	"net/http"

	"axial/models"
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

	if len(req.Users) == 0 {
		http.Error(w, "Users are required", http.StatusBadRequest)
		return
	}

	if err := ingest(models.DB, req.Users); err != nil {
		http.Error(w, "Failed to ingest users", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
