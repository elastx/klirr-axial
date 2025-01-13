package discovery

import (
	"fmt"
	"net"
)

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}

	panic("Unable to determine local IP")
}

func isUsableInterface(iface net.Interface) bool {
	// Skip interfaces that are:
	if iface.Flags&net.FlagUp == 0 {  // Not up
			return false
	}
	if iface.Flags&net.FlagLoopback != 0 {  // Loopback
			return false
	}
	if iface.Flags&net.FlagPointToPoint != 0 {  // Point-to-point
			return false
	}
	if iface.Flags&net.FlagMulticast == 0 {  // Not multicast capable
			return false
	}

	// Check if interface has any usable addresses
	addrs, err := iface.Addrs()
	if err != nil {
			return false
	}
	
	for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
					continue
			}
			if ipNet.IP.To4() != nil && !ipNet.IP.IsLoopback() {
					return true  // Found a usable IPv4 address
			}
	}
	
	return false  // No usable addresses found
}

func findMulticastInterface() (*net.Interface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
			return nil, fmt.Errorf("failed to get network interfaces: %v", err)
	}

	for _, iface := range ifaces {
			if isUsableInterface(iface) {
					fmt.Printf("Found interface %s with addresses: %v\n", iface.Name, getInterfaceAddrs(iface))
					return &iface, nil
			}
	}
	return nil, fmt.Errorf("no suitable multicast interface found")
}

func getInterfaceAddrs(iface net.Interface) []string {
	addrs, err := iface.Addrs()
	if err != nil {
			return []string{"error getting addresses"}
	}
	var result []string
	for _, addr := range addrs {
			result = append(result, addr.String())
	}
	return result
}