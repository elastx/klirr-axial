package api

import (
	"axial/config"
	"axial/models"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Global config reference
var appConfig *config.Config

// SetConfig initializes the config for the files API
func SetConfig(cfg *config.Config) {
	appConfig = cfg
}

// handleUploadFile handles multipart file uploads
func handleUploadFile(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (max memory 32MB)
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	// Get the file from form
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to retrieve file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get optional metadata
	description := r.FormValue("description")
	uploader := r.FormValue("uploader") // Fingerprint of uploader
	encryptedStr := r.FormValue("encrypted")
	recipientsStr := r.FormValue("recipients") // Comma-separated fingerprints

	if uploader == "" {
		http.Error(w, "Uploader fingerprint is required", http.StatusBadRequest)
		return
	}

	// Check max file size
	maxSize := appConfig.MaxFileSize
	if maxSize == 0 {
		maxSize = 100 * 1024 * 1024 // Default 100MB
	}

	if handler.Size > maxSize {
		http.Error(w, fmt.Sprintf("File size exceeds maximum of %d bytes", maxSize), http.StatusBadRequest)
		return
	}

	// Calculate content hash
	hasher := sha256.New()
	size, err := io.Copy(hasher, file)
	if err != nil {
		http.Error(w, "Failed to read file content", http.StatusInternalServerError)
		return
	}
	contentHash := hex.EncodeToString(hasher.Sum(nil))

	// Reset file pointer
	file.Seek(0, 0)

	// Parse encrypted flag
	encrypted := encryptedStr == "true"

	// Parse recipients
	var recipients models.Fingerprints
	if recipientsStr != "" {
		for _, fp := range strings.Split(recipientsStr, ",") {
			fp = strings.TrimSpace(fp)
			if fp != "" {
				recipients = append(recipients, models.Fingerprint(fp))
			}
		}
	}

	// Create file model
	fileModel := models.File{
		Uploader: models.Fingerprint(uploader),
		CreateFile: models.CreateFile{
			Filename:    handler.Filename,
			Size:        size,
			ContentType: handler.Header.Get("Content-Type"),
			Description: description,
			ContentHash: contentHash,
			Encrypted:   encrypted,
			Recipients:  recipients,
		},
	}

	// Save metadata to database
	if err := models.DB.Create(&fileModel).Error; err != nil {
		http.Error(w, fmt.Sprintf("Failed to save file metadata: %v", err), http.StatusInternalServerError)
		return
	}

	// Ensure storage directory exists
	storageRoot := appConfig.FileStoragePath
	if storageRoot == "" {
		storageRoot = "./data/files"
	}
	fullPath := filepath.Join(storageRoot, fileModel.StoragePath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		http.Error(w, "Failed to create storage directory", http.StatusInternalServerError)
		return
	}

	// Save file to disk
	outFile, err := os.Create(fullPath)
	if err != nil {
		http.Error(w, "Failed to create file on disk", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()

	file.Seek(0, 0)
	if _, err := io.Copy(outFile, file); err != nil {
		http.Error(w, "Failed to save file content", http.StatusInternalServerError)
		return
	}

	log.Printf("File uploaded: %s (ID: %s, Size: %d bytes)", fileModel.Filename, fileModel.ID, fileModel.Size)

	// Return file metadata
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fileModel)
}

// handleDownloadFile handles file downloads
func handleDownloadFile(w http.ResponseWriter, r *http.Request) {
	// Get file ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}
	fileID := pathParts[len(pathParts)-1]

	// Get file metadata from database
	var fileModel models.File
	if err := models.DB.First(&fileModel, "id = ?", fileID).Error; err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Open file from disk
	storageRoot := appConfig.FileStoragePath
	if storageRoot == "" {
		storageRoot = "./data/files"
	}
	fullPath := filepath.Join(storageRoot, fileModel.StoragePath)

	file, err := os.Open(fullPath)
	if err != nil {
		http.Error(w, "File not found on disk", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Set headers
	w.Header().Set("Content-Type", fileModel.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileModel.Filename))
	w.Header().Set("Content-Length", strconv.FormatInt(fileModel.Size, 10))

	// Stream file to response
	io.Copy(w, file)
}

// handleGetFiles lists files with optional filtering
func handleGetFiles(w http.ResponseWriter, r *http.Request) {
	query := models.DB.Model(&models.File{})

	// Filter by uploader
	if uploader := r.URL.Query().Get("uploader"); uploader != "" {
		query = query.Where("uploader = ?", uploader)
	}

	// Filter by content hash
	if contentHash := r.URL.Query().Get("content_hash"); contentHash != "" {
		query = query.Where("content_hash = ?", contentHash)
	}

	// Pagination
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	var files []models.File
	result := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&files)
	if result.Error != nil {
		http.Error(w, "Failed to retrieve files", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

// handleGetFileMetadata returns metadata for a specific file
func handleGetFileMetadata(w http.ResponseWriter, r *http.Request) {
	// Get file ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}
	fileID := pathParts[len(pathParts)-1]

	var fileModel models.File
	if err := models.DB.First(&fileModel, "id = ?", fileID).Error; err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fileModel)
}

// handleDeleteFile deletes a file (metadata and content)
func handleDeleteFile(w http.ResponseWriter, r *http.Request) {
	// Get file ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}
	fileID := pathParts[len(pathParts)-1]

	// Get file metadata
	var fileModel models.File
	if err := models.DB.First(&fileModel, "id = ?", fileID).Error; err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Delete file from disk
	storageRoot := appConfig.FileStoragePath
	if storageRoot == "" {
		storageRoot = "./data/files"
	}
	fullPath := filepath.Join(storageRoot, fileModel.StoragePath)
	if err := os.Remove(fullPath); err != nil {
		log.Printf("Warning: Failed to delete file from disk: %v", err)
	}

	// Delete metadata from database
	if err := models.DB.Delete(&fileModel).Error; err != nil {
		http.Error(w, "Failed to delete file metadata", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
