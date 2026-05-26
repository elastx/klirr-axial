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
		nodeID, _ = os.Hostname()
	}

	db, err := models.InitDB(cfg.Database)
	if err != nil {
		panic(fmt.Errorf("failed to initialize database: %v", err))
	}

	fmt.Printf("Starting node %s\n", nodeID)

	if err := models.RefreshHashes(db); err != nil {
		panic(fmt.Errorf("failed to calculate database hash: %v", err))
	}

	hashes := models.GetHashes()
	fmt.Printf("Node %s hash: %s\n", nodeID, hashes.Full)

	connections, err := discovery.CreateMulticastSockets(cfg)
	if err != nil {
		panic(err)
	}

	for _, conn := range connections {
		defer conn.Conn.Close()
		go discovery.StartMulticastListener(cfg, &conn, db)
		go discovery.StartBroadcast(cfg, &conn)
	}

	api.NewServer(db).RegisterRoutes()

	port := 8080
	fmt.Printf("Server starting on port %d...\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
