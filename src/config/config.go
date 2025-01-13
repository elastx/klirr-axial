package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

func LoadConfig() (Config, error) {
	cfg := Config{}

	// Load from ./config.yaml
	file, err := os.Open("config.yaml")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Config file not found, using default values")
			hostname, err := os.Hostname()
			if err != nil {
				hostname = "no-hostname"
			}
			cfg = Config{
				NodeID:           fmt.Sprintf("axial-%s", hostname),
				MulticastAddress: "239.192.0.1",
				MulticastPort:    45678,
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
	
	fmt.Println("Config loaded from config.yaml:")
	fmt.Printf("%+v\n", cfg)

	return cfg, nil
}
