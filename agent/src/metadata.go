package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type FileMetadata struct {
	OriginalPath string    `json:"original_path"`
	OriginalName string    `json:"original_name"`
	CurrentName  string    `json:"current_name"`
	FileSize     int64     `json:"file_size"`
	DeletedAt    time.Time `json:"deleted_at"`
	FileType     string    `json:"file_type"`
}

func writeMetadata(dateDir string, metadata FileMetadata) error {
	metadataFile := filepath.Join(dateDir, "metadata.json")
	var existingMetadata []FileMetadata

	if _, err := os.Stat(metadataFile); !os.IsNotExist(err) {
		data, err := os.ReadFile(metadataFile)
		if err != nil {
			return err
		}
		json.Unmarshal(data, &existingMetadata)
	}

	existingMetadata = append(existingMetadata, metadata)
	data, err := json.MarshalIndent(existingMetadata, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(metadataFile, data, 0644)
}
