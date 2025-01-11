package main

import (
	"fmt"
	"os"

	"axial/api"
	"axial/discovery"
)

func main() {
    nodeID := os.Getenv("NODE_ID")
    dataFile := os.Getenv("DATA_FILE")

    fmt.Printf("Starting node %s with data %s\n", nodeID, dataFile)

    go discovery.StartMulticastListener(nodeID)
    go api.StartHTTPServer()

    select {}
}
