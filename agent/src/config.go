package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	RecycleBinDir string `json:"recycleBinDir"`
	NumWorkers    int    `json:"numWorkers"`
}

func readConfig(filePath string) (Config, error) {
	var config Config
	data, err := os.ReadFile(filePath)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(data, &config)
	return config, err
}
