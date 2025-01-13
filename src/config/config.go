package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

func LoadConfig() (Config, error) {
	cfg := Config{}

	// Load from ./config.yaml
	file, err := os.Open("config.yaml")
	if err != nil {
		if os.IsNotExist(err) {
			cfg = Config{
				NodeID:           "default",
				MulticastAddress: "239.192.0.1",
				MulticastPort:    9999,
				APIPort:          8080,
				LogLevel:         "info",
				DataFile:         "data.yaml",
			}
			return cfg, nil
		}
		return cfg, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		return cfg, err
	}
	
	return cfg, nil
}
