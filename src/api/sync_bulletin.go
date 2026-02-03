package api

import (
	"axial/models"
	"encoding/json"
	"net/http"
)

type SyncBulletinRequest struct {
	Posts []models.Bulletin `json:"posts"`
}

func handleSyncBulletin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SyncBulletinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Posts) == 0 {
		http.Error(w, "Posts are required", http.StatusBadRequest)
		return
	}

	for _, post := range req.Posts {
		if err := models.DB.Create(&post).Error; err != nil {
			if models.IsDuplicateError(err) {
				continue
			}
			http.Error(w, "Failed to create bulletin", http.StatusInternalServerError)
			return
		}
	}

	models.RefreshHashes(models.DB)
	w.WriteHeader(http.StatusCreated)
}
