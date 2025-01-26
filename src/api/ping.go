package api

import (
	"encoding/json"
	"net/http"

	"axial/database"
	"axial/models"
)

type PingResponse struct {
	Hash string `json:"hash"`
	IsBusy bool `json:"is_busy"`
}

func handlePing(w http.ResponseWriter, _ *http.Request) {
	hash, err := models.GetDatabaseHash(database.DB)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	isSyncing := models.IsSyncing()

	response := PingResponse{
		Hash: hash,
		IsBusy: isSyncing,
	}

	json.NewEncoder(w).Encode(response)
}