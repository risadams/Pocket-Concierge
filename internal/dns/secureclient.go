package dns

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/miekg/dns"
	"github.com/risadams/Pocket-Concierge/internal/config"
)

// SecureClient handles secure DNS protocols
type SecureClient struct {
	httpClient *http.Client
	tlsConfig  *tls.Config
}

// NewSecureClient creates a new secure DNS client
func NewSecureClient() *SecureClient {
	return &SecureClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
				},
			},
		},
		tlsConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}
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

// queryTraditional handles UDP/TCP DNS queries
func (sc *SecureClient) queryTraditional(msg *dns.Msg, upstream config.UpstreamServer) (*dns.Msg, error) {
	client := &dns.Client{
		Net:     upstream.Protocol,
		Timeout: 5 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", upstream.Address, upstream.Port)
	response, _, err := client.Exchange(msg, addr)
	return response, err
}

// queryDoT handles DNS-over-TLS queries
func (sc *SecureClient) queryDoT(msg *dns.Msg, upstream config.UpstreamServer) (*dns.Msg, error) {
	client := &dns.Client{
		Net:     "tcp-tls",
		Timeout: 10 * time.Second,
		TLSConfig: &tls.Config{
			ServerName:         upstream.Address,
			InsecureSkipVerify: !upstream.Verify,
			MinVersion:         tls.VersionTLS12,
		},
	}

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
