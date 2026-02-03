package config

type DatabaseConfig struct {
	Host     string `yaml:"host" env:"DB_HOST"`
	Port     int    `yaml:"port" env:"DB_PORT"`
	User     string `yaml:"user" env:"DB_USER"`
	Password string `yaml:"password" env:"DB_PASSWORD"`
	Name     string `yaml:"name" env:"DB_NAME"`
}

type Config struct {
	NodeID           string         `args:"--node-id" yaml:"node_id" env:"NODE_ID"`
	MulticastAddress string         `args:"--multicast-address" yaml:"multicast_address" env:"MULTICAST_ADDRESS"`
	MulticastPort    int            `args:"--multicast-port" yaml:"multicast_port" env:"MULTICAST_PORT"`
	APIPort          int            `args:"--api-port" yaml:"api_port" env:"API_PORT"`
	LogLevel         string         `args:"--log-level" yaml:"log_level" env:"LOG_LEVEL"`
	FileStoragePath  string         `args:"--file-storage-path" yaml:"file_storage_path" env:"FILE_STORAGE_PATH"`
	MaxFileSize      int64          `args:"--max-file-size" yaml:"max_file_size" env:"MAX_FILE_SIZE"` // in bytes
	Database         DatabaseConfig `yaml:"database"`
}
