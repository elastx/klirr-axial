package api

import (
	"encoding/json"
	"net/http"

	"axial/models"
)

type PingResponse struct {
	Hashes models.HashSet `json:"hash"`
	IsBusy bool           `json:"is_busy"`
}

func (s *Server) handlePing(w http.ResponseWriter, _ *http.Request) {
	hashes, err := models.GetDatabaseHashes(s.DB)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := PingResponse{
		Hashes: hashes,
		IsBusy: models.IsSyncing(),
	}

	json.NewEncoder(w).Encode(response)
}
