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

func getNetworkInfo() {
	ifaces, err := net.Interfaces()
	if err != nil {
			fmt.Printf("Error getting interfaces: %v\n", err)
			return
	}
	
	for _, iface := range ifaces {
			addrs, err := iface.Addrs()
			if err != nil {
					continue
			}
			fmt.Printf("Interface: %v\n", iface.Name)
			for _, addr := range addrs {
					fmt.Printf("  Addr: %v\n", addr)
			}
	}
}