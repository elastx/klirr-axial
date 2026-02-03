package synchronization

import (
	"axial/models"
	"axial/remote"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// SyncFiles synchronizes file metadata with a remote node
func SyncFiles(node remote.API, files []models.File) error {
	endpoint := node.SyncFiles()
	responseData, response, err := endpoint.Post(files)
	if err != nil {
		return err
	}

	fmt.Printf("Received %s response from %s: %+v\n", response.Status, node.Address, responseData)

	return nil
}

// DownloadFile downloads a file from a remote node
func DownloadFile(node remote.API, fileID string, storageRoot string) (*models.File, error) {
	// First, get file metadata
	metadataEndpoint := node.FileMetadata(fileID)
	
	fileModel, response, err := metadataEndpoint.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %v", err)
	}
	
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get file metadata: status %d", response.StatusCode)
	}

	// Check if file already exists locally
	var existingFile models.File
	result := models.DB.First(&existingFile, "id = ?", fileModel.ID)
	if result.Error == nil {
		// File already exists, check if content matches
		fullPath := filepath.Join(storageRoot, existingFile.StoragePath)
		if _, err := os.Stat(fullPath); err == nil {
			log.Printf("File %s already exists locally", fileModel.Filename)
			return &existingFile, nil
		}
	}
	if result.Error == nil {
		// File already exists, check if content matches
		fullPath := filepath.Join(storageRoot, existingFile.StoragePath)
		if _, err := os.Stat(fullPath); err == nil {
			log.Printf("File %s already exists locally", fileModel.Filename)
			return &existingFile, nil
		}
	}

	// Download file content
	downloadURL := fmt.Sprintf("http://%s/v1/files/download/%s", node.Address, fileID)
	resp, err := http.Get(downloadURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file: status %d", resp.StatusCode)
	}

	// Ensure storage directory exists
	fullPath := filepath.Join(storageRoot, fileModel.StoragePath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %v", err)
	}

	// Save file to disk
	outFile, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file on disk: %v", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return nil, fmt.Errorf("failed to save file content: %v", err)
	}

	// Save metadata to database
	if result.Error != nil {
		// File doesn't exist, create it
		if err := models.DB.Create(&fileModel).Error; err != nil {
			return nil, fmt.Errorf("failed to save file metadata: %v", err)
		}
	} else {
		// File exists but content was missing, update it
		existingFile.StoragePath = fileModel.StoragePath
		if err := models.DB.Save(&existingFile).Error; err != nil {
			return nil, fmt.Errorf("failed to update file metadata: %v", err)
		}
		return &existingFile, nil
	}

	log.Printf("Downloaded file: %s (ID: %s, Size: %d bytes)", fileModel.Filename, fileModel.ID, fileModel.Size)
	return &fileModel, nil
}

// SyncMissingFiles downloads files that exist on remote node but not locally
func SyncMissingFiles(node remote.API, remoteFiles []models.File, storageRoot string) error {
	for _, remoteFile := range remoteFiles {
		// Check if file exists locally
		var localFile models.File
		result := models.DB.First(&localFile, "id = ?", remoteFile.ID)
		
		if result.Error != nil {
			// File doesn't exist locally, download it
			log.Printf("Downloading missing file: %s (ID: %s)", remoteFile.Filename, remoteFile.ID)
			_, err := DownloadFile(node, remoteFile.ID, storageRoot)
			if err != nil {
				log.Printf("Failed to download file %s: %v", remoteFile.Filename, err)
				continue
			}
		}
	}
	
	return nil
}
