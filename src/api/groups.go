package api

import (
	"axial/models"
	"encoding/json"
	"net/http"
)

func handleGetGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var groups []models.Group
	if err := models.DB.Find(&groups).Error; err != nil {
		http.Error(w, "Failed to fetch groups", http.StatusInternalServerError)
		return
	}

	// Collect all fingerprints needed to hydrate in one query
	fpSet := make(map[string]struct{})
	for _, g := range groups {
		if g.UserID != "" {
			fpSet[g.UserID] = struct{}{}
		}
		for _, m := range g.Members {
			s := string(m)
			if s != "" {
				fpSet[s] = struct{}{}
			}
		}
	}

	// Build slice of unique fingerprints
	fps := make([]string, 0, len(fpSet))
	for fp := range fpSet {
		fps = append(fps, fp)
	}

	// Fetch all users matching those fingerprints
	usersByFP := make(map[string]models.User, len(fps))
	if len(fps) > 0 {
		var users []models.User
		if err := models.DB.Where("fingerprint IN ?", fps).Find(&users).Error; err != nil {
			http.Error(w, "Failed to fetch users for groups", http.StatusInternalServerError)
			return
		}
		for _, u := range users {
			usersByFP[u.Fingerprint] = u
		}
	}

	// Assemble hydrated groups
	hydrated := make([]models.HydratedGroup, 0, len(groups))
	for _, g := range groups {
		hg := models.HydratedGroup{Group: g}
		if u, ok := usersByFP[g.UserID]; ok {
			hg.User = u
		}
		if len(g.Members) > 0 {
			users := make([]models.User, 0, len(g.Members))
			for _, m := range g.Members {
				if u, ok := usersByFP[string(m)]; ok {
					users = append(users, u)
				}
			}
			hg.Users = users
		}
		hydrated = append(hydrated, hg)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hydrated)
}

func handleGetGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	var group models.Group
	if err := models.DB.Where("id = ?", id).First(&group).Error; err != nil {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// Hydrate single group
	hg := models.HydratedGroup{Group: group}

	// Fetch owner user if exists
	if group.UserID != "" {
		var owner models.User
		if err := models.DB.Where("fingerprint = ?", group.UserID).First(&owner).Error; err == nil {
			hg.User = owner
		}
	}

	// Fetch member users that exist
	if len(group.Members) > 0 {
		// Convert to []string for IN query
		mfp := make([]string, 0, len(group.Members))
		for _, f := range group.Members {
			if s := string(f); s != "" {
				mfp = append(mfp, s)
			}
		}
		if len(mfp) > 0 {
			var members []models.User
			if err := models.DB.Where("fingerprint IN ?", mfp).Find(&members).Error; err == nil {
				hg.Users = members
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hg)
}

func handleRegisterGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reg models.CreateGroup
	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	group := models.Group{
		CreateGroup: reg,
	}

	if err := models.DB.Create(&group).Error; err != nil {
		http.Error(w, "Failed to register group", http.StatusInternalServerError)
		return
	}

	models.RefreshHashes(models.DB)

	w.WriteHeader(http.StatusCreated)
} 

// Alias for router: create group via POST /v1/groups
func handleCreateGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reg models.CreateGroup
	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	group := models.Group{CreateGroup: reg}

	if err := models.DB.Create(&group).Error; err != nil {
		http.Error(w, "Failed to register group", http.StatusInternalServerError)
		return
	}

	models.RefreshHashes(models.DB)

	w.WriteHeader(http.StatusCreated)
}
