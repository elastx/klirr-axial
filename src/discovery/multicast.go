package discovery

import (
	"fmt"
	"net"
)

func StartMulticastListener(nodeID string) {
	addr, err := net.ResolveUDPAddr("udp", "224.0.0.1:9999")
	if err != nil {
			panic(err)
	}

	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
			panic(err)
	}
	defer conn.Close()

	buffer := make([]byte, 1024)
	for {
			n, src, err := conn.ReadFromUDP(buffer)
			if err != nil {
					fmt.Println("Error receiving multicast:", err)
					continue
			}

			fmt.Printf("Received from %s: %s\n", src, string(buffer[:n]))
	}
}

func SendMulticast(nodeID string, hash string, addr string) {
	conn, err := net.Dial("udp", "224.0.0.1:9999")
	if err != nil {
			panic(err)
	}
	defer conn.Close()

	message := fmt.Sprintf("%s|%s|%s", nodeID, hash, addr)
	conn.Write([]byte(message))
}
