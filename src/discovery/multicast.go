package discovery

import (
	"fmt"
	"net"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/ipv4"

	"axial/config"
	"axial/database"
	"axial/models"
	"axial/synchronization"
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

	fmt.Printf("Attempting to bind to all interfaces first (port %d)\n", cfg.MulticastPort)
	// First try binding to ALL interfaces
	if conn, err := setupMulticastConn(cfg, nil, addr); err == nil {
		fmt.Printf("Successfully bound to all interfaces\n")
		connections = append(connections, MulticastConnection{
			Conn:    conn,
			iface:   nil,
			localIP: "0.0.0.0",
		})
		return connections, nil  // Return early if we successfully bound to all interfaces
	} else {
		fmt.Printf("Failed to bind to all interfaces: %v\n", err)
		// Only try individual interfaces if binding to all interfaces failed
		for _, iface := range ifaces {
			if !isUsableInterface(iface) {
				fmt.Printf("Skipping interface %s (not usable)\n", iface.Name)
				continue
			}

			fmt.Printf("Attempting to bind to interface %s (port %d)\n", iface.Name, cfg.MulticastPort)
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
	}

	if len(connections) == 0 {
		return nil, fmt.Errorf("no usable interfaces found")
	}

	return connections, nil
}

func isBroadcast(ip net.IP) bool {
	// Check for 255.255.255.255 or x.x.x.255
	if ip == nil {
			return false
	}
	if ip.To4() == nil {
			return false
	}
	return ip[len(ip)-1] == 255
}


func setupMulticastConn(cfg config.Config, iface *net.Interface, addr *net.UDPAddr) (*net.UDPConn, error) {
	var conn *net.UDPConn
	var err error

	// Check if we're using broadcast
	if isBroadcast(addr.IP) {
		// For broadcast, bind to 0.0.0.0
		laddr := &net.UDPAddr{
			IP:   net.IPv4zero,
			Port: cfg.MulticastPort,
		}
		conn, err = net.ListenUDP("udp4", laddr)
		if err != nil {
			return nil, fmt.Errorf("failed to create broadcast socket: %v", err)
		}

		// Set broadcast permission
		rawConn, err := conn.SyscallConn()
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to get raw connection: %v", err)
		}

		err = rawConn.Control(func(fd uintptr) {
			err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
			if err != nil {
				panic(fmt.Errorf("failed to set SO_BROADCAST: %v", err))
			}
			// Also set SO_REUSEADDR
			err = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
			if err != nil {
				panic(fmt.Errorf("failed to set SO_REUSEADDR: %v", err))
			}
		})
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to set socket options: %v", err)
		}
	} else {

		conn, err := net.ListenMulticastUDP("udp4", iface, addr)
		if err != nil {
			return nil, fmt.Errorf("failed to create socket: %v. Addr: %v", err, addr)
		}

		p := ipv4.NewPacketConn(conn)

		if iface != nil {
			// Set the interface for outgoing multicast traffic
			if err := p.SetMulticastInterface(iface); err != nil {
				conn.Close()
				return nil, fmt.Errorf("SetMulticastInterface failed: %v", err)
			}
		}

		// Disable multicast loopback
		if err := p.SetMulticastLoopback(false); err != nil {
			conn.Close()
			return nil, fmt.Errorf("SetMulticastLoopback failed: %v", err)
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

		// Check if MulticastAddress is an IP address or hostname
		multicastIP := net.ParseIP(cfg.MulticastAddress)
		if multicastIP == nil {
			// Resolve hostname to IP address
			multicastIPs, err := net.LookupIP(cfg.MulticastAddress)
			if err != nil {
				conn.Close()
				return nil, fmt.Errorf("failed to resolve multicast address: %v", err)
			}
			multicastIP = multicastIPs[0]
		}

		// Force leave and rejoin of the multicast group
		if err := p.LeaveGroup(iface, &net.UDPAddr{IP: multicastIP}); err != nil {
			fmt.Printf("Warning: failed to leave multicast group: %v\n", err)
		}

		if err := p.JoinGroup(iface, &net.UDPAddr{IP: multicastIP}); err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to join multicast group: %v", err)
		}

		if iface == nil {
			fmt.Printf("Successfully joined multicast group %s on all interfaces\n", multicastIP)
		} else {
			fmt.Printf("Successfully joined multicast group %s on interface %s\n", multicastIP, iface.Name)
		}

		// Set buffer sizes
		if err := conn.SetReadBuffer(1048576); err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to set read buffer: %v", err)
		}

		if err := conn.SetWriteBuffer(1048576); err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to set write buffer: %v", err)
		}
	}
	return conn, nil
}

func StartMulticastListener(cfg config.Config, conn *MulticastConnection) {
	fmt.Printf("Listening for messages on %v\n", conn.Conn.LocalAddr())
	buffer := make([]byte, 4096)

	for {
		n, src, err := conn.Conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error receiving message:", err)
			continue
		}

		message := string(buffer[:n])

		// Check for both our configured port and the Mac's port (60090)
		if !strings.Contains(src.String(), fmt.Sprintf(":%d", cfg.MulticastPort)) {
			if strings.Contains(src.String(), ":60090") {
				fmt.Printf("Got message on port 60090 from %s: %q\n", src, message)
			} else {
				fmt.Printf("Message on unexpected port from %s\n", src)
			}
		}

		// Only process messages that look like ours (4 pipe-separated fields)
		if parts := strings.Split(message, "|"); len(parts) == 4 {
			fmt.Printf("RECV: %s (from %s)\n", message, src)
			// axial-mix.local|74d63e48f0e18e7c300904b49457a630ec782c244fb212273742ce1499cd21ef|:8080|0.0.0.0 (from 192.168.1.207:45678)
			if false && !models.IsSyncing() {
				hash := parts[1]
				ourHash, err := models.GetDatabaseHash(database.DB)
				if err != nil {
					fmt.Printf("Failed to get database hash: %v\n", err)
					continue
				}

				if hash != ourHash {
					fmt.Printf("Mismatching hash from %s: %s != %s\n", src, hash, ourHash)
					port := parts[2]
					remoteNode := models.RemoteNode{
						Hash: hash,
						Address: fmt.Sprintf("%s%s", src.IP, port),
					}
					
					err := synchronization.StartSync(remoteNode)
					if err != nil {
						fmt.Printf("Failed to start sync: %v\n", err)
					} else {
						fmt.Printf("Synchronized with %s\n", remoteNode.Address)
					}
				} else {
					fmt.Printf("Matching hash from %s\n", src)
				}
			} else {
				fmt.Printf("Ignoring ping from %s because we're already syncing\n", src)
			}

		} else {
			// Debug log for non-matching messages
			fmt.Printf("Ignored non-axial message from %s (len=%d)\n", src, len(message))
		}
	}
}

func StartBroadcast(cfg config.Config, hash string, conn *MulticastConnection) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	addr := fmt.Sprintf(":%d", cfg.APIPort)
	message := fmt.Sprintf("%s|%s|%s|%s", cfg.NodeID, hash, addr, conn.localIP)

	targetAddr := net.UDPAddr{
		IP:   net.IPv4(255, 255, 255, 255),
		Port: cfg.MulticastPort,
	}

	fmt.Printf("Starting broadcast from %s to %s:%d\n", conn.localIP, targetAddr.IP, targetAddr.Port)

	for range ticker.C {
		_, err := conn.Conn.WriteToUDP([]byte(message), &targetAddr)
		if err != nil {
			fmt.Printf("Error sending broadcast message: %v\n", err)
		} else {
			fmt.Printf("SENT: %s\n", message)
		}
	}
}
