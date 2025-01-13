package discovery

import (
	"fmt"
	"net"
	"syscall"
	"time"

	"golang.org/x/net/ipv4"

	"axial/config"
)

func CreateMulticastSocket(cfg config.Config) (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", cfg.MulticastPort))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address: %v", err)
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to create socket: %v", err)
	}

	// Set socket options using raw fd
	rawConn, err := conn.SyscallConn()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to get raw connection: %v", err)
	}

	err = rawConn.Control(func(fd uintptr) {
		// Check if MulticastAddress is an IP address or hostname
		multicastIP := net.ParseIP(cfg.MulticastAddress)
		if multicastIP == nil {
			// Resolve hostname to IP address
			multicastIPs, err := net.LookupIP(cfg.MulticastAddress)
			if err != nil {
				panic(fmt.Errorf("failed to resolve multicast address: %v", err))
			}
			multicastIP = multicastIPs[0]
		}

		// Join multicast group
		p := ipv4.NewPacketConn(conn)
		ifaces, err := net.Interfaces()
		if err != nil {
			panic(fmt.Errorf("failed to get network interfaces: %v", err))
		}

		var iface *net.Interface
		for _, i := range ifaces {
			if i.Flags&net.FlagMulticast != 0 {
				iface = &i
				break
			}
		}

		if iface == nil {
			panic(fmt.Errorf("no multicast-enabled interface found"))
		}

		err = p.JoinGroup(iface, &net.UDPAddr{IP: multicastIP})
		if err != nil {
			panic(fmt.Errorf("failed to join multicast group: %v", err))
		}

		// Set multicast TTL
		err = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_MULTICAST_TTL, 2)
		if err != nil {
			panic(fmt.Errorf("failed to set multicast TTL: %v", err))
		}
	})
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to set socket options: %v", err)
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

func StartMulticastListener(cfg config.Config, conn *net.UDPConn) {
	fmt.Printf("Listening for messages on %v\n", conn.LocalAddr())
	buffer := make([]byte, 4096)

	for {
		n, src, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error receiving message:", err)
			continue
		}
		message := string(buffer[:n])
		fmt.Printf("Node %s received from %s: %s\n", cfg.NodeID, src, message)
	}
}

func StartBroadcast(cfg config.Config, hash string, addr string, conn *net.UDPConn) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	localIP := getLocalIP()

	// Resolve multicast address to an IP address
	targetAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:9999", cfg.MulticastAddress))
	if err != nil {
		fmt.Printf("Error resolving multicast address: %v\n", err)
		return
	}

	fmt.Printf("Starting broadcast from %s to %s\n", localIP, targetAddr)

	for range ticker.C {
		message := fmt.Sprintf("%s|%s|%s|%s\n", cfg.NodeID, hash, addr, localIP)
		_, err := conn.WriteToUDP([]byte(message), targetAddr)
		if err != nil {
			fmt.Printf("Error sending multicast message: %v\n", err)
		} else {
			fmt.Printf("Broadcasted: node=%s hash=%s addr=%s ip=%s\n", cfg.NodeID, hash, addr, localIP)
		}
	}
}
