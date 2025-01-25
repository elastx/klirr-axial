package api

import (
	"encoding/json"
	"net/http"

	"axial/database"
	"axial/models"
)

type MessageRequest struct {
	Topic       string `json:"topic"`
	Content     string `json:"content"`
	Fingerprint string `json:"fingerprint"`
}

func handleGetMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var messages []models.Message
	if err := database.DB.Find(&messages).Error; err != nil {
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

	var req MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Topic == "" || req.Content == "" || req.Fingerprint == "" {
		http.Error(w, "Topic, content and fingerprint are required", http.StatusBadRequest)
		return
	}

	// Verify user exists
	var user models.User
	if err := database.DB.Where("fingerprint = ?", req.Fingerprint).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	message := models.Message{
		Topic:       req.Topic,
		Content:     req.Content,
		Fingerprint: req.Fingerprint,
	}

	if err := database.DB.Create(&message).Error; err != nil {
		http.Error(w, "Failed to create message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func handleGetTopics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var topics []string
	if err := database.DB.Model(&models.Message{}).Distinct().Pluck("topic", &topics).Error; err != nil {
		http.Error(w, "Failed to fetch topics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(topics)
} 