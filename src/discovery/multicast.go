package discovery

import (
	"fmt"
	"net"
	"syscall"
	"time"

	"golang.org/x/net/ipv4"

	"axial/config"
)

// New type to hold our connections
type MulticastConnection struct {
	Conn    *net.UDPConn
	iface   *net.Interface
	localIP string
}

func CreateMulticastSockets(cfg config.Config) ([]MulticastConnection, error) {
	var connections []MulticastConnection

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get interfaces: %v", err)
	}

	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", cfg.MulticastAddress, cfg.MulticastPort))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address: %v", err)
	}

	// Try to create a connection for each usable interface
	for _, iface := range ifaces {
		if !isUsableInterface(iface) {
			continue
		}

		conn, err := setupMulticastConn(cfg, &iface, addr)
		if err != nil {
			fmt.Printf("Warning: failed to setup multicast on interface %s: %v\n", iface.Name, err)
			continue
		}

		// Get the local IP for this interface
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		var localIP string
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok {
				if ip4 := ipNet.IP.To4(); ip4 != nil && !ip4.IsLoopback() {
					localIP = ip4.String()
					break
				}
			}
		}

		connections = append(connections, MulticastConnection{
			Conn:    conn,
			iface:   &iface,
			localIP: localIP,
		})

		fmt.Printf("Successfully joined multicast group on interface %s (IP: %s)\n",
			iface.Name, localIP)
	}

	if len(connections) == 0 {
		return nil, fmt.Errorf("no usable interfaces found")
	}

	return connections, nil
}

func setupMulticastConn(cfg config.Config, iface *net.Interface, addr *net.UDPAddr) (*net.UDPConn, error) {
	conn, err := net.ListenMulticastUDP("udp4", iface, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to create socket: %v. Addr: %v", err, addr)
	}

	// Set socket options using raw fd
	rawConn, err := conn.SyscallConn()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to get raw connection: %v", err)
	}

	err = rawConn.Control(func(fd uintptr) {
		// Allow multiple sockets to use the same port
		err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
		if err != nil {
			panic(fmt.Errorf("failed to set SO_REUSEADDR: %v", err))
		}

		// Allow sending to loopback interface
		err = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_MULTICAST_LOOP, 1)
		if err != nil {
			panic(fmt.Errorf("failed to set IP_MULTICAST_LOOP: %v", err))
		}

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

		fmt.Printf("Joining multicast group %s on interface %s\n", multicastIP, iface.Name)

		err = p.SetMulticastInterface(iface)
		if err != nil {
			panic(fmt.Errorf("failed to set multicast interface: %v", err))
		}

		// Leave the multicast group if already joined
		err = p.LeaveGroup(iface, &net.UDPAddr{IP: multicastIP})
		if err != nil {
			fmt.Printf("Failed to leave multicast group: %v\n", err)
		}

		err = p.JoinGroup(iface, &net.UDPAddr{IP: multicastIP})
		if err != nil {
			panic(fmt.Errorf("failed to join multicast group: %v. Args: %+v %+v", err, iface, &net.UDPAddr{IP: multicastIP}))
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
			
			fmt.Printf("Received %d bytes from %s\n", n, src)
			fmt.Printf("Raw bytes: %v\n", buffer[:n])
			fmt.Printf("As string: %q\n", buffer[:n])
	}
}

func StartBroadcast(cfg config.Config, hash string, addr string, conn *MulticastConnection) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Start with a simple test message
	testMessage := []byte("test\n")

	targetAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", cfg.MulticastAddress, cfg.MulticastPort))
	if err != nil {
		fmt.Printf("Error resolving multicast address: %v\n", err)
		return
	}

	fmt.Printf("Starting broadcast to %s\n", targetAddr)

	for range ticker.C {
		_, err := conn.Conn.WriteToUDP(testMessage, targetAddr)
		if err != nil {
			fmt.Printf("Error sending multicast message: %v\n", err)
		} else {
			fmt.Printf("Broadcasted test message\n")
			// Print the raw bytes we're sending
			fmt.Printf("Raw bytes sent: %v\n", testMessage)
		}
	}
}
