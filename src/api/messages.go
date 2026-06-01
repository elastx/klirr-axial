package api

import (
	"encoding/json"
	"log"
	"net/http"

	"axial/models"
)

func (s *Server) handleGetMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var messages []models.Message
	if err := s.DB.Find(&messages).Error; err != nil {
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func (s *Server) handleCreateMessage(w http.ResponseWriter, r *http.Request) {
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

	if err := s.DB.Create(&message).Error; err != nil {
		log.Printf("Create message failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	models.RefreshHashes(s.DB)

	w.WriteHeader(http.StatusCreated)
}
