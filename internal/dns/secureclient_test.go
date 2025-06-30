package dns

import (
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/risadams/Pocket-Concierge/internal/config"
)

func TestNewSecureClient(t *testing.T) {
	client := NewSecureClient()
	if client == nil {
		t.Fatal("NewSecureClient returned nil")
	}

	if client.httpClient == nil {
		t.Error("HTTP client not initialized")
	}

	if client.tlsConfig == nil {
		t.Error("TLS config not initialized")
	}

	if client.clients == nil {
		t.Error("DNS clients map not initialized")
	}

	// Check HTTP client timeout
	if client.httpClient.Timeout != 5*time.Second {
		t.Errorf("Expected HTTP timeout 5s, got %v", client.httpClient.Timeout)
	}
}

func TestSecureClientQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping secure client test in short mode")
	}

	client := NewSecureClient()

	// Create a test query
	query := &dns.Msg{}
	query.SetQuestion("google.com.", dns.TypeA)

	tests := []struct {
		name     string
		upstream config.UpstreamServer
		skipTest bool
		reason   string
	}{
		{
			name: "UDP upstream",
			upstream: config.UpstreamServer{
				Address:  "8.8.8.8",
				Protocol: "udp",
				Port:     53,
				Verify:   false,
			},
			skipTest: false,
		},
		{
			name: "TCP upstream",
			upstream: config.UpstreamServer{
				Address:  "8.8.8.8",
				Protocol: "tcp",
				Port:     53,
				Verify:   false,
			},
			skipTest: false,
		},
		{
			name: "DoT upstream",
			upstream: config.UpstreamServer{
				Address:  "1.1.1.1",
				Protocol: "tls",
				Port:     853,
				Verify:   false,
			},
			skipTest: true, // Skip DoT test as it requires working TLS
			reason:   "DoT requires network access and working TLS",
		},
		{
			name: "DoH upstream",
			upstream: config.UpstreamServer{
				Address:  "cloudflare-dns.com",
				Protocol: "https",
				Port:     443,
				Path:     "/dns-query",
				Verify:   false,
			},
			skipTest: true, // Skip DoH test as it requires network access
			reason:   "DoH requires network access",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipTest {
				t.Skip(tt.reason)
			}

			// For basic UDP/TCP tests, check if we can reach external DNS
			if (tt.upstream.Protocol == "udp" || tt.upstream.Protocol == "tcp") && !canReachExternalDNS() {
				t.Skip("Cannot reach external DNS servers")
			}

			response, err := client.Query(query, tt.upstream)

			// We expect either a successful response or a timeout/network error
			// Don't fail the test for network issues in CI environments
			if err != nil {
				t.Logf("Query failed (expected in some environments): %v", err)
				return
			}

			if response == nil {
				t.Error("Expected response but got nil")
				return
			}

			if response.Id != query.Id {
				t.Errorf("Response ID mismatch: expected %d, got %d", query.Id, response.Id)
			}

			// For google.com, we should get some answers
			if len(response.Answer) == 0 {
				t.Log("No answers returned (may be expected in test environment)")
			}
		})
	}
}

func TestSecureClientInvalidUpstream(t *testing.T) {
	client := NewSecureClient()

	query := &dns.Msg{}
	query.SetQuestion("test.com.", dns.TypeA)

	tests := []struct {
		name     string
		upstream config.UpstreamServer
	}{
		{
			name: "invalid protocol",
			upstream: config.UpstreamServer{
				Address:  "8.8.8.8",
				Protocol: "invalid",
				Port:     53,
			},
		},
		{
			name: "unreachable server",
			upstream: config.UpstreamServer{
				Address:  "192.0.2.1", // RFC5737 test address
				Protocol: "udp",
				Port:     53,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := client.Query(query, tt.upstream)

			// Should either return an error or nil response
			if err == nil && response != nil {
				t.Error("Expected error or nil response for invalid upstream")
			}
		})
	}
}

func TestSecureClientConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrency test in short mode")
	}

	if !canReachExternalDNS() {
		t.Skip("Cannot reach external DNS servers")
	}

	client := NewSecureClient()

	upstream := config.UpstreamServer{
		Address:  "8.8.8.8",
		Protocol: "udp",
		Port:     53,
		Verify:   false,
	}

	// Test concurrent queries
	const numGoroutines = 10
	const queriesPerGoroutine = 5

	resultChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(workerID int) {
			for j := 0; j < queriesPerGoroutine; j++ {
				query := &dns.Msg{}
				query.SetQuestion("google.com.", dns.TypeA)

				_, err := client.Query(query, upstream)
				if err != nil {
					resultChan <- err
					return
				}
			}
			resultChan <- nil
		}(i)
	}

	// Wait for all goroutines to complete
	errorCount := 0
	for i := 0; i < numGoroutines; i++ {
		err := <-resultChan
		if err != nil {
			errorCount++
			t.Logf("Concurrent query error: %v", err)
		}
	}

	// Allow some errors in CI environments, but not all
	if errorCount == numGoroutines {
		t.Error("All concurrent queries failed")
	}
}

func TestSecureClientConnectionPooling(t *testing.T) {
	client := NewSecureClient()

	// Create multiple clients for the same upstream
	upstream := config.UpstreamServer{
		Address:  "8.8.8.8",
		Protocol: "udp",
		Port:     53,
	}

	// Get DNS client multiple times
	dnsClient1 := client.getOrCreateClient(upstream)
	dnsClient2 := client.getOrCreateClient(upstream)

	// Should return the same client instance (connection pooling)
	if dnsClient1 != dnsClient2 {
		t.Error("Expected same DNS client instance for connection pooling")
	}

	// Test with different upstream
	upstream2 := config.UpstreamServer{
		Address:  "1.1.1.1",
		Protocol: "udp",
		Port:     53,
	}

	dnsClient3 := client.getOrCreateClient(upstream2)

	// Should return different client for different upstream
	if dnsClient1 == dnsClient3 {
		t.Error("Expected different DNS client instance for different upstream")
	}
}

func TestSecureClientTLSConfig(t *testing.T) {
	tests := []struct {
		name     string
		upstream config.UpstreamServer
		checkTLS bool
	}{
		{
			name: "TLS with verification disabled",
			upstream: config.UpstreamServer{
				Address:  "1.1.1.1",
				Protocol: "tls",
				Port:     853,
				Verify:   false,
			},
			checkTLS: true,
		},
		{
			name: "TLS with verification enabled",
			upstream: config.UpstreamServer{
				Address:  "9.9.9.9", // Use different address to avoid cache collision
				Protocol: "tls",
				Port:     853,
				Verify:   true,
			},
			checkTLS: true,
		},
		{
			name: "UDP - no TLS",
			upstream: config.UpstreamServer{
				Address:  "8.8.8.8",
				Protocol: "udp",
				Port:     53,
			},
			checkTLS: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh client for each test to avoid cache interference
			client := NewSecureClient()
			dnsClient := client.getOrCreateClient(tt.upstream)

			if tt.checkTLS {
				// For TLS clients, verify TLS config is set
				if dnsClient.TLSConfig == nil {
					t.Error("Expected TLS config for TLS client")
				} else {
					// When Verify=true, InsecureSkipVerify should be false (secure)
					// When Verify=false, InsecureSkipVerify should be true (insecure)
					expectedInsecureSkipVerify := !tt.upstream.Verify
					if dnsClient.TLSConfig.InsecureSkipVerify != expectedInsecureSkipVerify {
						t.Errorf("Expected InsecureSkipVerify=%v when Verify=%v, got %v",
							expectedInsecureSkipVerify, tt.upstream.Verify, dnsClient.TLSConfig.InsecureSkipVerify)
					}
				}
			} else {
				// For non-TLS clients, TLS config should be nil
				if dnsClient.TLSConfig != nil {
					t.Error("Expected no TLS config for non-TLS client")
				}
			}
		})
	}
}

func BenchmarkSecureClientQuery(b *testing.B) {
	if !canReachExternalDNS() {
		b.Skip("Cannot reach external DNS servers")
	}

	client := NewSecureClient()

	upstream := config.UpstreamServer{
		Address:  "8.8.8.8",
		Protocol: "udp",
		Port:     53,
		Verify:   false,
	}

	query := &dns.Msg{}
	query.SetQuestion("google.com.", dns.TypeA)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.Query(query, upstream)
		if err != nil {
			b.Logf("Query failed: %v", err)
		}
	}
}

func BenchmarkSecureClientConcurrent(b *testing.B) {
	if !canReachExternalDNS() {
		b.Skip("Cannot reach external DNS servers")
	}

	client := NewSecureClient()

	upstream := config.UpstreamServer{
		Address:  "8.8.8.8",
		Protocol: "udp",
		Port:     53,
		Verify:   false,
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			query := &dns.Msg{}
			query.SetQuestion("google.com.", dns.TypeA)

			_, err := client.Query(query, upstream)
			if err != nil {
				// Don't fail benchmark for network errors
			}
		}
	})
}

// Helper function to check external DNS connectivity
func canReachExternalDNS() bool {
	client := &dns.Client{Timeout: 3 * time.Second}
	query := &dns.Msg{}
	query.SetQuestion("google.com.", dns.TypeA)

	_, _, err := client.Exchange(query, "8.8.8.8:53")
	return err == nil
}
