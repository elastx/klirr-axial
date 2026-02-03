package api

import (
	"axial/models"
	"encoding/json"
	"net/http"
)

type UserRegistration struct {
	PublicKey   string `json:"public_key"`
}

func handleGetUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var users []models.User
	if err := models.DB.Find(&users).Error; err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fingerprint := r.PathValue("fingerprint")
	if fingerprint == "" {
		http.Error(w, "Fingerprint is required", http.StatusBadRequest)
		return
	}

	var user models.User
	if err := models.DB.Where("fingerprint = ?", fingerprint).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func handleRegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reg UserRegistration
	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user := models.User{
		CreateUser: models.CreateUser{
			PublicKey: reg.PublicKey,
		},
	}

	if err := models.DB.Create(&user).Error; err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	models.RefreshHashes(models.DB)

	w.WriteHeader(http.StatusCreated)
} 