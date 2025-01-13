package discovery

import (
	"fmt"
	"net"
	"os"
	"time"
)

func CreateMulticastSocket() (*net.UDPConn, error) {
	// Bind to a local address and port to listen for UDP messages
	addr, err := net.ResolveUDPAddr("udp4", ":9999")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address: %v", err)
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to create socket: %v", err)
	}

	if err := conn.SetReadBuffer(1048576); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to set read buffer: %v", err)
	}

	if err := conn.SetWriteBuffer(1048576); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to set write buffer: %v", err)
	}

	return conn, nil
}

func StartMulticastListener(nodeID string, conn *net.UDPConn) {
	fmt.Printf("Listening for messages on %v\n", conn.LocalAddr())
	buffer := make([]byte, 4096)

	for {
		n, src, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error receiving message:", err)
			continue
		}
		message := string(buffer[:n])
		fmt.Printf("Node %s received from %s: %s\n", nodeID, src, message)
	}
}

func StartBroadcast(nodeID string, hash string, addr string, conn *net.UDPConn) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	localIP := getLocalIP()

	// Use multicast address from environment variable, fallback to relay service
	multicastAddress := os.Getenv("MULTICAST_ADDRESS")
	if multicastAddress == "" {
		multicastAddress = "multicast-relay.default.svc.cluster.local"
	}

	// Resolve multicast address to an IP address
	targetAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:9999", multicastAddress))
	if err != nil {
		fmt.Printf("Error resolving multicast address: %v\n", err)
		return
	}

	fmt.Printf("Starting broadcast from %s to %s\n", localIP, targetAddr)

	for range ticker.C {
		message := fmt.Sprintf("%s|%s|%s|%s\n", nodeID, hash, addr, localIP)
		_, err := conn.WriteToUDP([]byte(message), targetAddr)
		if err != nil {
			fmt.Printf("Error sending multicast message: %v\n", err)
		} else {
			fmt.Printf("Broadcasted: node=%s hash=%s addr=%s ip=%s\n", nodeID, hash, addr, localIP)
		}
	}
}
