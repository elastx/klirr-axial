package data

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

func LoadData(filePath string) ([]DataBlock, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Create the file if it does not exist
		err = CreateDataFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create file %s: %v", filePath, err)
		}
	}

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
	var rawData struct {
		Data []DataBlock `yaml:"data"`
	}
	err = yaml.Unmarshal(bytes, &rawData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML from file %s: %v", filePath, err)
	}

	return rawData.Data, nil
}

func CreateDataFile(filePath string) error {
	data := struct {
		Data []DataBlock `yaml:"data"`
	}{
		Data: []DataBlock{},
	}

	yamlBytes, err := yaml.Marshal(&data)
	if err != nil {
		return fmt.Errorf("failed to marshal data to YAML: %v", err)
	}

	// Write to file
	err = os.WriteFile(filePath, yamlBytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %v", filePath, err)
	}

	return nil
}
