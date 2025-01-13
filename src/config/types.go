package config

type Config struct {
	NodeID           string `args:"--node-id" yaml:"node_id" env:"NODE_ID"`
	MulticastAddress string `args:"--multicast-address" yaml:"multicast_address" env:"MULTICAST_ADDRESS"`
	MulticastPort    int    `args:"--multicast-port" yaml:"multicast_port" env:"MULTICAST_PORT"`
	APIPort          int    `args:"--api-port" yaml:"api_port" env:"API_PORT"`
	LogLevel         string `args:"--log-level" yaml:"log_level" env:"LOG_LEVEL"`
	DataFile         string `args:"--data-file" yaml:"data_file" env:"DATA_FILE"`
}
