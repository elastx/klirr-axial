package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/axial/")
	viper.AddConfigPath("$HOME/.axial")

	envConfigPath := viper.GetString("CONFIG_PATH")
	if envConfigPath != "" {
		viper.AddConfigPath(envConfigPath)
	}

	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	viper.AutomaticEnv()

	viper.SetDefault("NodeID", "axial")
	viper.SetDefault("Port", 8080)
	viper.SetDefault("LogLevel", "info")
	viper.SetDefault("MulticastAddress", "239.192.0.1")
	viper.SetDefault("MulticastPort", 9999)

	if err := viper.ReadInConfig(); err != nil {
		log.Println("Error reading config file, using defaults and env variables")
	}
}

func LoadConfig() Config {
	return Config{
		NodeID:  viper.GetString("NodeID"),
		APIPort:     viper.GetInt("Port"),
		LogLevel: viper.GetString("LogLevel"),
		MulticastAddress: viper.GetString("MulticastAddress"),
		MulticastPort: viper.GetInt("MulticastPort"),
	}
}