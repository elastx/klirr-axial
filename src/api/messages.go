package api

import (
	"encoding/json"
	"net/http"

	"axial/models"
)


func handleGetMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var messages []models.Message
	if err := models.DB.Find(&messages).Error; err != nil {
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func handleCreateMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateMessage
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	message := models.Message{
		CreateMessage: req,
	}

	if err := models.DB.Create(&message).Error; err != nil {
		http.Error(w, "Failed to create message", http.StatusInternalServerError)
		return
	}

	models.RefreshHashes(models.DB)

	w.WriteHeader(http.StatusCreated)
}
