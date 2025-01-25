package main

import (
	"fmt"
	"os"

	"axial/api"
	"axial/config"
	"axial/data"
	"axial/database"
	"axial/discovery"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	nodeID := cfg.NodeID
	if nodeID == "" {
		// Default to hostname
		nodeID, _ = os.Hostname()
	}

	// Initialize database connection
	err = database.Connect(cfg.Database)
	if err != nil {
		panic(fmt.Errorf("failed to initialize database: %v", err))
	}

	dataFile := cfg.DataFile
	fmt.Printf("Starting node %s with data %s\n", nodeID, dataFile)

	// Load data
	loadedData, err := data.LoadData(dataFile)
	if err != nil {
		panic(err)
	}

	convertedData := make([]data.DataBlock, len(loadedData))
	for i, v := range loadedData {
		convertedData[i] = data.DataBlock(v)
	}
	api.SetData(convertedData)

	// Calculate initial hash
	hash := data.CalculateHash(loadedData)
	fmt.Printf("Node %s hash: %s\n", nodeID, hash)

	// Create single multicast socket
	connections, err := discovery.CreateMulticastSockets(cfg)
	if err != nil {
		panic(err)
	}
	
	for _, conn := range connections {
		defer conn.Conn.Close()
		go discovery.StartMulticastListener(cfg, &conn)
		go discovery.StartBroadcast(cfg, hash, &conn)
	}
	go api.StartHTTPServer()

	select {}
}
