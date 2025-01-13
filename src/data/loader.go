package data

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type DataBlock struct {
	ID      string `json:"id"`
	User    string `json:"user"`
	Content string `json:"content"`
}

func LoadData(filePath string) ([]DataBlock, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %v", filePath, err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	// Parse YAML
	var rawData map[string]interface{}
	err = yaml.Unmarshal(bytes, &rawData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML from file %s: %v", filePath, err)
	}

	// Extract "data" key and marshal it to JSON
	dataKey, ok := rawData["data"]
	if !ok {
		return nil, fmt.Errorf("missing 'data' key in file %s", filePath)
	}

	jsonBytes, err := json.Marshal(dataKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data to JSON: %v", err)
	}

	// Parse JSON into DataBlock array
	var data []DataBlock
	err = json.Unmarshal(jsonBytes, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON data: %v", err)
	}

	return data, nil
}
