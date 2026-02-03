package api

import (
	"axial/models"
	"encoding/json"
	"net/http"
)

type SyncFilesRequest struct {
	Files []models.File `json:"files"`
}

func handleSyncFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SyncFilesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if len(req.Files) == 0 {
		http.Error(w, "Files are required", http.StatusBadRequest)
		return
	}

	// Create file metadata entries
	// Note: This only syncs metadata, not the actual file content
	// File content should be downloaded separately using the file download endpoint
	for _, file := range req.Files {
		if err := models.DB.Create(&file).Error; err != nil {
			// Ignore duplicate errors
			if models.IsDuplicateError(err) {
				continue
			}
			http.Error(w, "Failed to create file metadata", http.StatusInternalServerError)
			return
		}
	}

	models.RefreshHashes(models.DB)

	w.WriteHeader(http.StatusCreated)
}
