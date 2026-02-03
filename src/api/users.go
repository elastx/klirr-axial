package api

import (
	"axial/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"sort"
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

// GET /v1/users/search?q=...&limit=20&offset=0
func handleSearchUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	offset := 0
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil {
			limit = v
		}
	}
	if offsetStr != "" {
		if v, err := strconv.Atoi(offsetStr); err == nil {
			offset = v
		}
	}
	if limit <= 0 { limit = 20 }
	if limit > 50 { limit = 50 }
	if offset < 0 { offset = 0 }

	// Validate query length: allow short hex-like prefixes, else require >=2 chars
	isHex := func(s string) bool {
		if len(s) < 16 { return false }
		for _, c := range s {
			if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F') {
				return false
			}
		}
		return true
	}
	if q == "" || (!isHex(q) && len(q) < 2) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"users": []models.User{}, "next_offset": nil})
		return
	}

	var users []models.User
	var total int64
	// Fingerprint substring match (case-insensitive)
	if err := models.DB.Model(&models.User{}).
		Where("fingerprint ILIKE ?", "%"+q+"%").
		Count(&total).Error; err != nil {
		http.Error(w, "Failed to count users", http.StatusInternalServerError)
		return
	}
	if err := models.DB.
		Where("fingerprint ILIKE ?", "%"+q+"%").
		Order("fingerprint ASC").
		Limit(limit).Offset(offset).
		Find(&users).Error; err != nil {
		http.Error(w, "Failed to search users", http.StatusInternalServerError)
		return
	}

	var nextOffset *int
	if int64(offset+limit) < total {
		no := offset + limit
		nextOffset = &no
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"users": users, "next_offset": nextOffset})
}

// GET /v1/users/recent?limit=10
// Derive distinct counterpart fingerprints from messages involving the current user.
func handleRecentUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil { limit = v }
	}
	if limit <= 0 { limit = 10 }
	if limit > 50 { limit = 50 }

	current := strings.TrimSpace(r.Header.Get("X-User-Fingerprint"))
	if current == "" {
		http.Error(w, "X-User-Fingerprint header required", http.StatusBadRequest)
		return
	}

	// Fetch messages where the user is sender or among recipients (JSONB array contains)
	var messages []models.Message
	arrayContains := fmt.Sprintf("[\"%s\"]", current)
	if err := models.DB.
		Where("sender = ?", current).
		Or("to_jsonb(recipients)::jsonb @> ?", arrayContains).
		Order("created_at DESC").
		Find(&messages).Error; err != nil {
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}

	// Compute last interaction time per counterpart
	type stamp struct { t time.Time }
	last := map[string]stamp{}
	for _, m := range messages {
		cur := string(m.Sender)
		// If current is sender, counterparts are recipients
		if cur == current {
			for _, rfp := range m.Recipients {
				fp := string(rfp)
				if fp == current { continue }
				if s, ok := last[fp]; !ok || m.CreatedAt.After(s.t) {
					last[fp] = stamp{t: m.CreatedAt}
				}
			}
		} else {
			// Otherwise counterpart is sender, if current appears in recipients
			found := false
			for _, rfp := range m.Recipients {
				if string(rfp) == current { found = true; break }
			}
			if found {
				fp := cur
				if s, ok := last[fp]; !ok || m.CreatedAt.After(s.t) {
					last[fp] = stamp{t: m.CreatedAt}
				}
			}
		}
	}

	// Order counterparts by last timestamp
	type pair struct { fp string; t time.Time }
	pairs := make([]pair, 0, len(last))
	for fp, s := range last { pairs = append(pairs, pair{fp: fp, t: s.t}) }
	// Sort descending by time
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].t.After(pairs[j].t) })
	if len(pairs) > limit { pairs = pairs[:limit] }

	// Fetch users for these fingerprints, preserving order
	fps := make([]string, 0, len(pairs))
	for _, p := range pairs { fps = append(fps, p.fp) }
	var users []models.User
	if len(fps) > 0 {
		if err := models.DB.Where("fingerprint IN ?", fps).Find(&users).Error; err != nil {
			http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
			return
		}
	}
	// Map for quick lookup
	umap := map[string]models.User{}
	for _, u := range users { umap[u.Fingerprint] = u }
	ordered := make([]models.User, 0, len(fps))
	for _, fp := range fps {
		if u, ok := umap[fp]; ok { ordered = append(ordered, u) }
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ordered)
}