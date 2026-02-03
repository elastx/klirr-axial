package models

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// CreateFile represents the data required to create a file entry
type CreateFile struct {
	Filename    string `json:"filename" gorm:"column:filename;not null"`
	Size        int64  `json:"size" gorm:"column:size;not null"`
	ContentType string `json:"content_type" gorm:"column:content_type;not null"`
	Description string `json:"description,omitempty" gorm:"column:description"`
	// ContentHash is SHA-256 of the actual file content (not PGP data)
	ContentHash string `json:"content_hash" gorm:"column:content_hash;not null"`
	// Encrypted indicates if the file is PGP encrypted
	Encrypted bool `json:"encrypted" gorm:"column:encrypted;default:false"`
	// Recipients list (if encrypted)
	Recipients Fingerprints `json:"recipients,omitempty" gorm:"column:recipients;type:jsonb"`
}

// File represents a file in the system
type File struct {
	Base
	Uploader Fingerprint `json:"uploader" gorm:"column:uploader;not null;index"`
	CreateFile
	// StoragePath is the local filesystem path (relative to storage root)
	StoragePath string `json:"storage_path" gorm:"column:storage_path;not null"`
	// Signature of the file metadata (optional)
	Signature string `json:"signature,omitempty" gorm:"column:signature"`
}

func (f *File) In(files []File) bool {
	for _, file := range files {
		if f.ID == file.ID {
			return true
		}
	}
	return false
}

// TableName specifies the table name for File
func (File) TableName() string {
	return "files"
}

// Hash creates a deterministic file ID based on file properties
func (f *File) Hash() string {
	// Create a string combining all relevant properties
	recipients := ""
	for _, r := range f.Recipients {
		recipients += string(r)
	}
	
	hashStrings := []string{
		string(f.Uploader),
		f.Filename,
		fmt.Sprintf("%d", f.Size),
		f.ContentType,
		f.ContentHash,
		f.Description,
		recipients,
		f.CreatedAt.Format(time.RFC3339Nano),
	}

	idBytes := []byte{}
	for _, s := range hashStrings {
		idBytes = append(idBytes, []byte(s)...)
	}

	// Generate SHA-256 hash
	hash := sha256.Sum256([]byte(idBytes))

	// Convert to hex string
	return hex.EncodeToString(hash[:])
}

func (f *File) BeforeCreate(*gorm.DB) error {
	// Validate file size (max 100MB by default)
	maxSize := int64(100 * 1024 * 1024)
	if f.Size > maxSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxSize)
	}

	// Validate filename
	if f.Filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	// Validate content hash
	if f.ContentHash == "" {
		return fmt.Errorf("content hash cannot be empty")
	}

	// If encrypted, recipients must be specified
	if f.Encrypted && len(f.Recipients) == 0 {
		return fmt.Errorf("encrypted files must have at least one recipient")
	}

	// Generate storage path based on content hash (first 2 chars for directory structure)
	if f.StoragePath == "" && f.ContentHash != "" {
		f.StoragePath = fmt.Sprintf("%s/%s/%s", 
			f.ContentHash[0:2], 
			f.ContentHash[2:4], 
			f.ContentHash)
	}

	f.Base.BeforeCreate(f.Hash())

	return nil
}

// GetByContentHash retrieves files by their content hash
func GetFilesByContentHash(db *gorm.DB, contentHash string) ([]File, error) {
	var files []File
	result := db.Where("content_hash = ?", contentHash).Find(&files)
	return files, result.Error
}

// GetFilesByUploader retrieves all files uploaded by a specific fingerprint
func GetFilesByUploader(db *gorm.DB, uploader Fingerprint) ([]File, error) {
	var files []File
	result := db.Where("uploader = ?", uploader).Order("created_at DESC").Find(&files)
	return files, result.Error
}

// GetRecentFiles retrieves the most recent files
func GetRecentFiles(db *gorm.DB, limit int) ([]File, error) {
	var files []File
	result := db.Order("created_at DESC").Limit(limit).Find(&files)
	return files, result.Error
}
