package main

import (
	"fmt"
	"os"

	"axial/api"
	"axial/data"
	"axial/discovery"
)

func main() {
    nodeID := os.Getenv("NODE_ID")
    dataFile := os.Getenv("DATA_FILE")

    fmt.Printf("Starting node %s with data %s\n", nodeID, dataFile)

    // Load data
    loadedData, err := data.LoadData(dataFile)
    if err != nil {
        panic(err)
    }
    convertedData := make([]api.DataBlock, len(loadedData))
    for i, v := range loadedData {
        convertedData[i] = api.DataBlock(v)
    }
    api.SetData(convertedData)

    // Calculate initial hash
    hash := data.CalculateHash(loadedData)
    fmt.Printf("Node %s hash: %s\n", nodeID, hash)

    // Create single multicast socket
    conn, err := discovery.CreateMulticastSocket()
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    // Start services
    go discovery.StartMulticastListener(nodeID, conn)
    go discovery.StartBroadcast(nodeID, hash, ":8080", conn)
    go api.StartHTTPServer()

    select {}
}