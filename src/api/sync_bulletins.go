package api

import (
	"encoding/json"
	"net/http"

	"axial/models"
)

type SyncBulletinsRequest struct {
	Bulletins []models.Bulletin `json:"bulletins"`
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

	if len(req.Bulletins) == 0 {
		http.Error(w, "Bulletins are required", http.StatusBadRequest)
		return
	}

	if err := ingest(models.DB, req.Bulletins); err != nil {
		http.Error(w, "Failed to ingest bulletins", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
