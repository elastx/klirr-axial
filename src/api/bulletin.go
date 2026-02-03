package api

import (
	"encoding/json"
	"net/http"
	"log"

	"axial/models"
)

func handleGetBulletin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var posts []models.Bulletin
	if err := models.DB.Order("created_at DESC").Find(&posts).Error; err != nil {
		http.Error(w, "Failed to fetch bulletin posts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func handleCreateBulletin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateBulletin
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	post := models.Bulletin{
		CreateBulletin: req,
	}

	if err := models.DB.Create(&post).Error; err != nil {
		log.Printf("Create bulletin failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	models.RefreshHashes(models.DB)

	w.WriteHeader(http.StatusCreated)
}
