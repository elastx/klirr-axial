package api

import (
	"encoding/json"
	"net/http"

	"axial/models"
)

type SyncUsersRequest struct {
	Users []models.User `json:"users"`
}

func (s *Server) handleSyncUsers(w http.ResponseWriter, r *http.Request) {
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

	for _, user := range req.Users {
		if err := s.DB.Create(&user).Error; err != nil {
			if models.IsDuplicateError(err) {
				continue
			}
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
	}

	models.RefreshHashes(s.DB)

	w.WriteHeader(http.StatusCreated)
}
