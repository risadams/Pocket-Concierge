package dns

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/risadams/Pocket-Concierge/internal/config"
)

// SecureClient handles secure DNS protocols with optimized connection pooling
type SecureClient struct {
	httpClient   *http.Client
	tlsConfig    *tls.Config
	clients      map[string]*dns.Client
	clientsMutex sync.RWMutex
}

// NewSecureClient creates a new secure DNS client with optimized settings
func NewSecureClient() *SecureClient {
	// Configure optimized HTTP transport
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	return &SecureClient{
		httpClient: &http.Client{
			Timeout:   5 * time.Second, // Reduced from 10s
			Transport: transport,
		},
		tlsConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		clients: make(map[string]*dns.Client),
	}
}

// getOrCreateClient returns a cached client for the given upstream server
func (sc *SecureClient) getOrCreateClient(upstream config.UpstreamServer) *dns.Client {
	key := fmt.Sprintf("%s:%s:%d", upstream.Protocol, upstream.Address, upstream.Port)

	sc.clientsMutex.RLock()
	if client, exists := sc.clients[key]; exists {
		sc.clientsMutex.RUnlock()
		return client
	}
	sc.clientsMutex.RUnlock()

	sc.clientsMutex.Lock()
	defer sc.clientsMutex.Unlock()

	// Double-check after acquiring write lock
	if client, exists := sc.clients[key]; exists {
		return client
	}

	// Create new optimized client
	var client *dns.Client

	switch upstream.Protocol {
	case "udp", "tcp":
		client = &dns.Client{
			Net:     upstream.Protocol,
			Timeout: 3 * time.Second, // Reduced timeout
		}
	case "tls":
		client = &dns.Client{
			Net:     "tcp-tls",
			Timeout: 5 * time.Second,
			TLSConfig: &tls.Config{
				ServerName:         upstream.Address,
				InsecureSkipVerify: !upstream.Verify,
				MinVersion:         tls.VersionTLS12,
			},
		}
	default:
		client = &dns.Client{
			Net:     "udp",
			Timeout: 3 * time.Second,
		}
	}

	sc.clients[key] = client
	return client
}

// Query sends a DNS query using the specified upstream server
func (sc *SecureClient) Query(msg *dns.Msg, upstream config.UpstreamServer) (*dns.Msg, error) {
	switch upstream.Protocol {
	case "udp", "tcp":
		return sc.queryTraditional(msg, upstream)
	case "tls":
		return sc.queryDoT(msg, upstream)
	case "https":
		return sc.queryDoH(msg, upstream)
	case "quic":
		return nil, fmt.Errorf("DNS-over-QUIC not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", upstream.Protocol)
	}
}

// queryTraditional handles UDP/TCP DNS queries with connection pooling
func (sc *SecureClient) queryTraditional(msg *dns.Msg, upstream config.UpstreamServer) (*dns.Msg, error) {
	client := sc.getOrCreateClient(upstream)
	addr := fmt.Sprintf("%s:%d", upstream.Address, upstream.Port)
	response, _, err := client.Exchange(msg, addr)
	return response, err
}

// queryDoT handles DNS-over-TLS queries with connection pooling
func (sc *SecureClient) queryDoT(msg *dns.Msg, upstream config.UpstreamServer) (*dns.Msg, error) {
	client := sc.getOrCreateClient(upstream)
	addr := fmt.Sprintf("%s:%d", upstream.Address, upstream.Port)
	response, _, err := client.Exchange(msg, addr)
	return response, err
}

// queryDoH handles DNS-over-HTTPS queries
func (sc *SecureClient) queryDoH(msg *dns.Msg, upstream config.UpstreamServer) (*dns.Msg, error) {
	// Pack DNS message to wire format
	packed, err := msg.Pack()
	if err != nil {
		return nil, fmt.Errorf("failed to pack DNS message: %w", err)
	}

	// Create HTTPS request
	url := fmt.Sprintf("https://%s:%d%s", upstream.Address, upstream.Port, upstream.Path)
	req, err := http.NewRequest("POST", url, bytes.NewReader(packed))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set DoH headers
	req.Header.Set("Content-Type", "application/dns-message")
	req.Header.Set("Accept", "application/dns-message")
	req.Header.Set("User-Agent", "PocketConcierge/1.0")

	// Send request
	resp, err := sc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Unpack DNS response
	response := &dns.Msg{}
	if err := response.Unpack(body); err != nil {
		return nil, fmt.Errorf("failed to unpack DNS response: %w", err)
	}

	return response, nil
}
