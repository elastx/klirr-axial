package api

import (
	"encoding/json"
	"net/http"

	"axial/models"
)

type SyncMessagesRequest struct {
	Messages []models.Message `json:"messages"`
}

func (s *Server) handleSyncMessages(w http.ResponseWriter, r *http.Request) {
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

	for _, message := range req.Messages {
		if err := s.DB.Create(&message).Error; err != nil {
			if models.IsDuplicateError(err) {
				continue
			}
			http.Error(w, "Failed to create message", http.StatusInternalServerError)
			return
		}
	}

	models.RefreshHashes(s.DB)

	w.WriteHeader(http.StatusCreated)
}
