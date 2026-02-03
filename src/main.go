package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"axial/api"
	"axial/config"
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
	err = models.InitDB(cfg.Database)
	if err != nil {
		panic(fmt.Errorf("failed to initialize database: %v", err))
	}

	fmt.Printf("Starting node %s\n", nodeID)

	// Calculate initial hash
	err = models.RefreshHashes(models.DB)
	if err != nil {
		panic(fmt.Errorf("failed to calculate database hash: %v", err))
	}

	hashes := models.GetHashes()

	fmt.Printf("Node %s hash: %s\n", nodeID, hashes.Full)

	// Initialize API config (needed for file uploads)
	api.SetConfig(&cfg)

	// Create single multicast socket
	connections, err := discovery.CreateMulticastSockets(cfg)
	if err != nil {
		panic(err)
	}
	
	for _, conn := range connections {
		defer conn.Conn.Close()
		go discovery.StartMulticastListener(cfg, &conn)
		go discovery.StartBroadcast(cfg, &conn)
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
