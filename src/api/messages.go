package api

import (
	"encoding/json"
	"net/http"

	"axial/database"
	"axial/models"
)

type MessageRequest struct {
	Topic       string  `json:"topic,omitempty"`
	Recipient   string  `json:"recipient,omitempty"`
	Content     string  `json:"content"`
	Fingerprint string  `json:"fingerprint"`
	ParentID    *string `json:"parent_id,omitempty"`
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
	if req.Content == "" || req.Fingerprint == "" {
		http.Error(w, "Content and fingerprint are required", http.StatusBadRequest)
		return
	}

	// Verify sender exists
	var sender models.User
	if err := database.DB.Where("fingerprint = ?", req.Fingerprint).First(&sender).Error; err != nil {
		http.Error(w, "Sender not found", http.StatusNotFound)
		return
	}

	messageType := "bulletin"
	if req.Recipient != "" {
		messageType = "private"
		// For private messages, verify recipient exists
		var recipient models.User
		if err := database.DB.Where("fingerprint = ?", req.Recipient).First(&recipient).Error; err != nil {
			http.Error(w, "Recipient not found", http.StatusNotFound)
			return
		}
	} else if req.Topic == "" {
		http.Error(w, "Either topic or recipient is required", http.StatusBadRequest)
		return
	}

	// If ParentID is provided, verify it exists and is a bulletin post
	if req.ParentID != nil {
		var parent models.Message
		if err := database.DB.First(&parent, "message_id = ?", *req.ParentID).Error; err != nil {
			http.Error(w, "Parent message not found", http.StatusNotFound)
			return
		}
		if parent.Type != "bulletin" {
			http.Error(w, "Can only reply to bulletin messages", http.StatusBadRequest)
			return
		}
		// Inherit topic from parent
		req.Topic = parent.Topic
	}

	message := models.Message{
		Topic:     req.Topic,
		Recipient: req.Recipient,
		Content:   req.Content,
		Author:    req.Fingerprint,
		Type:      messageType,
		ParentID:  req.ParentID,
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