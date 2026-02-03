package api

import (
	"encoding/json"
	"net/http"

	"axial/models"
)

type PingResponse struct {
	Hashes models.HashSet `json:"hash"`
	IsBusy bool `json:"is_busy"`
}

func handlePing(w http.ResponseWriter, _ *http.Request) {
	hashes, err := models.GetDatabaseHashes(models.DB)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	isSyncing := models.IsSyncing()

	response := PingResponse{
		Hashes: hashes,
		IsBusy: isSyncing,
	}

	json.NewEncoder(w).Encode(response)
}