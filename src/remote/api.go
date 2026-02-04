package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type API struct {
	Scheme  string
	Address string
	Port    int
}

type Endpoint[PostRequestType any, PostResponseType any, GetResponseType any] struct {
	Node              *API
	Version           string
	Path              string
	ValidPostResponse func(http.Response) bool
	ValidGetResponse  func(http.Response) bool
}

func (e *Endpoint[_, _, _]) URL() *url.URL {
	if e.Node == nil {
		return nil
	}
	return &url.URL{
		Scheme: e.Node.Scheme,
		Host:   fmt.Sprintf("%s:%d", e.Node.Address, e.Node.Port),
		Path:   e.Path,
	}
}

// isBlockedIP returns true if the IP is loopback, link-local, multicast, unspecified,
// or matches any of the local interface addresses (to avoid hitting services on this host).
func isBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	// Unspecified
	if ip.IsUnspecified() {
		return true
	}
	// Loopback
	if ip.IsLoopback() {
		return true
	}
	// Multicast
	if ip.IsMulticast() {
		return true
	}
	// Link-local IPv4 169.254.0.0/16
	if v4 := ip.To4(); v4 != nil {
		if v4[0] == 169 && v4[1] == 254 {
			return true
		}
	}
	// Link-local IPv6 fe80::/10
	if ip.To16() != nil && ip.To4() == nil {
		// Check fe80::/10
		// First 10 bits: 1111111010 -> fe80::/10
		b0 := ip[0]
		b1 := ip[1]
		if b0 == 0xfe && (b1&0xc0) == 0x80 {
			return true
		}
		// ::1 handled by IsLoopback above
	}

	// Block any address assigned to local interfaces
	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()
		for _, a := range addrs {
			var ipAddr net.IP
			switch v := a.(type) {
			case *net.IPNet:
				ipAddr = v.IP
			case *net.IPAddr:
				ipAddr = v.IP
			}
			if ipAddr != nil && ipAddr.Equal(ip) {
				return true
			}
		}
	}
	return false
}

// restrictedDial dials only to non-blocked IPs and enforces port policy.
func restrictedDial(ctx context.Context, network, address string, allowedPort int) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	// Enforce expected port (mitigates SSRF to arbitrary local services)
	if allowedPort > 0 && port != fmt.Sprint(allowedPort) {
		return nil, fmt.Errorf("blocked port: %s (expected %d)", port, allowedPort)
	}

	// Block obvious bad hosts
	lowerHost := strings.ToLower(host)
	if lowerHost == "localhost" {
		return nil, fmt.Errorf("blocked host: localhost")
	}

	// If host is an IP, check directly
	ip := net.ParseIP(host)
	if ip != nil {
		if isBlockedIP(ip) {
			return nil, fmt.Errorf("blocked IP address: %s", ip.String())
		}
		d := &net.Dialer{Timeout: 30 * time.Second}
		return d.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
	}

	// Resolve hostname and choose the first allowed IP
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	for _, a := range addrs {
		if a.IP == nil || isBlockedIP(a.IP) {
			continue
		}
		d := &net.Dialer{Timeout: 30 * time.Second}
		return d.DialContext(ctx, network, net.JoinHostPort(a.IP.String(), port))
	}
	return nil, fmt.Errorf("blocked: all resolved IPs are disallowed for host %s", host)
}

// safeHTTPClient builds an HTTP client that prevents SSRF to localhost/metadata and enforces port.
func safeHTTPClient(allowedPort int) *http.Client {
	// Transport with restricted dialers for both HTTP and HTTPS
	tr := &http.Transport{}
	tr.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		return restrictedDial(ctx, network, address, allowedPort)
	}
	// Enforce the same restriction for TLS
	tr.DialTLSContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		return restrictedDial(ctx, network, address, allowedPort)
	}
	return &http.Client{Transport: tr, Timeout: 30 * time.Second}
}

func (e *Endpoint[T, R, _]) Post(data T) (R, *http.Response, error) {
	var result R
	// Basic URL scheme validation
	url := e.URL()
	if url == nil {
		return result, nil, fmt.Errorf("invalid endpoint: no node configured")
	}
	if url.Scheme != "http" && url.Scheme != "https" {
		return result, nil, fmt.Errorf("unsupported URL scheme: %s", url.Scheme)
	}

	body, err := json.Marshal(data)
	if err != nil {
		return result, nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	client := safeHTTPClient(e.Node.Port)
	resp, err := client.Post(url.String(), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return result, resp, fmt.Errorf("failed to perform POST request: %w", err)
	}

	if e.ValidPostResponse != nil && !e.ValidPostResponse(*resp) {
		return result, resp, fmt.Errorf("invalid response: %s", resp.Status)
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, resp, fmt.Errorf("failed to decode response: %w", err)
	}
	return result, resp, nil
}

func (e *Endpoint[_, _, R]) Get() (R, *http.Response, error) {
	var result R
	// Basic URL scheme validation
	url := e.URL()
	if url == nil {
		return result, nil, fmt.Errorf("invalid endpoint: no node configured")
	}
	if url.Scheme != "http" && url.Scheme != "https" {
		return result, nil, fmt.Errorf("unsupported URL scheme: %s", url.Scheme)
	}
	client := safeHTTPClient(e.Node.Port)
	resp, err := client.Get(url.String())
	if err != nil {
		return result, resp, fmt.Errorf("failed to perform GET request: %w", err)
	}

	if e.ValidGetResponse != nil && !e.ValidGetResponse(*resp) {
		return result, resp, fmt.Errorf("invalid response: %s", resp.Status)
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, resp, fmt.Errorf("failed to decode response: %w", err)
	}
	return result, resp, nil
}
