package api

import (
	"axial/models"
	"encoding/json"
	"net/http"
)

type SyncBulletinsRequest struct {
	Bulletins []models.Bulletin `json:"messages"`
}

func handleSyncBulletins(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SyncBulletinsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if len(req.Bulletins) == 0 {
		http.Error(w, "Bulletins are required", http.StatusBadRequest)
		return
	}

	// Create bulletins
	for _, bulletin := range req.Bulletins {
		if err := models.DB.Create(&bulletin).Error; err != nil {
			// Ignore duplicate errors
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