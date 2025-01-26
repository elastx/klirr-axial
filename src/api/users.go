package api

import (
	"encoding/json"
	"net/http"

	"axial/database"
	"axial/models"
)

type UserRegistration struct {
	Fingerprint string `json:"fingerprint"`
	PublicKey   string `json:"public_key"`
}

func handleGetUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var users []models.User
	if err := database.DB.Find(&users).Error; err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
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

	// Validate required fields
	if reg.Fingerprint == "" || reg.PublicKey == "" {
		http.Error(w, "Fingerprint and public key are required", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := database.DB.Where("fingerprint = ?", reg.Fingerprint).First(&existingUser).Error; err == nil {
		http.Error(w, "User already registered", http.StatusConflict)
		return
	}

	user := models.User{
		Fingerprint: reg.Fingerprint,
		PublicKey:   reg.PublicKey,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
} 