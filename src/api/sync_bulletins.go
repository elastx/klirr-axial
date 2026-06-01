package api

import (
	"encoding/json"
	"net/http"

	"axial/models"
)

type SyncBulletinsRequest struct {
	Bulletins []models.Bulletin `json:"messages"`
}

func (s *Server) handleSyncBulletins(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SyncBulletinsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Bulletins) == 0 {
		http.Error(w, "Bulletins are required", http.StatusBadRequest)
		return
	}

	for _, bulletin := range req.Bulletins {
		if err := s.DB.Create(&bulletin).Error; err != nil {
			if models.IsDuplicateError(err) {
				continue
			}
			http.Error(w, "Failed to create bulletin", http.StatusInternalServerError)
			return
		}
	}

	models.RefreshHashes(s.DB)

	w.WriteHeader(http.StatusCreated)
}
