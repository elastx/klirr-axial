package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"axial/api"
	"axial/config"
	"axial/database"
	"axial/discovery"
	"axial/models"
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

	fmt.Printf("Starting node %s\n", nodeID)

	// Calculate initial hash
	hash, err := models.GetDatabaseHash(database.DB)
	if err != nil {
		panic(fmt.Errorf("failed to calculate database hash: %v", err))
	}
	models.UpdateHash(hash)
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

	// Register API routes
	api.RegisterRoutes()

	// Start server
	port := 8080
	fmt.Printf("Server starting on port %d...\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
