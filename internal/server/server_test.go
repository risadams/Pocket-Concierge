package server

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/risadams/Pocket-Concierge/internal/config"
)

func TestNewServer(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Address: "127.0.0.1",
			Port:    8053,
		},
		DNS: config.DNSConfig{
			TTL:             300,
			EnableRecursion: true,
		},
		HomeDNSDomain: "home",
	}

	server := NewServer(cfg)
	if server == nil {
		t.Fatal("NewServer returned nil")
	}

	if server.config != cfg {
		t.Error("Server config not set correctly")
	}

	if server.dnsHandler == nil {
		t.Error("DNS handler not initialized")
	}

	if server.server == nil {
		t.Error("DNS server not initialized")
	}

	expectedAddr := fmt.Sprintf("%s:%d", cfg.Server.Address, cfg.Server.Port)
	if server.server.Addr != expectedAddr {
		t.Errorf("Expected server address %s, got %s", expectedAddr, server.server.Addr)
	}

	if server.server.Net != "udp" {
		t.Errorf("Expected UDP protocol, got %s", server.server.Net)
	}

	// Check timeouts are set
	if server.server.ReadTimeout != 3*time.Second {
		t.Errorf("Expected read timeout 3s, got %v", server.server.ReadTimeout)
	}

	if server.server.WriteTimeout != 3*time.Second {
		t.Errorf("Expected write timeout 3s, got %v", server.server.WriteTimeout)
	}

	if server.server.UDPSize != 65535 {
		t.Errorf("Expected UDP size 65535, got %d", server.server.UDPSize)
	}
}

func TestServerCheckPort(t *testing.T) {
	tests := []struct {
		name        string
		address     string
		port        int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid port",
			address:     "127.0.0.1",
			port:        0, // Use 0 to let OS assign a free port
			expectError: false,
		},
		{
			name:        "port 53 (may require privileges)",
			address:     "127.0.0.1",
			port:        53,
			expectError: false, // Don't assume this will fail - depends on user privileges
			errorMsg:    "",
		},
		{
			name:        "invalid address",
			address:     "999.999.999.999",
			port:        8053,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Server: config.ServerConfig{
					Address: tt.address,
					Port:    tt.port,
				},
				DNS: config.DNSConfig{
					TTL:             300,
					EnableRecursion: true,
				},
				HomeDNSDomain: "home",
			}

			server := NewServer(cfg)
			err := server.checkPort()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestServerGetStats(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Address: "127.0.0.1",
			Port:    8053,
		},
		DNS: config.DNSConfig{
			TTL:             300,
			EnableRecursion: true,
		},
		Hosts: []config.HostEntry{
			{Hostname: "test1.home", IPv4: []string{"192.168.1.100"}},
			{Hostname: "test2.home", IPv4: []string{"192.168.1.101"}},
		},
		Upstream: []config.UpstreamServer{
			{Address: "8.8.8.8", Protocol: "udp", Port: 53},
		},
		HomeDNSDomain: "home",
	}

	server := NewServer(cfg)
	stats := server.GetStats()

	if stats == nil {
		t.Fatal("GetStats returned nil")
	}

	// Check required fields
	expectedFields := []string{"address", "port", "upstream_dns", "local_hosts", "ttl", "recursion"}
	for _, field := range expectedFields {
		if _, exists := stats[field]; !exists {
			t.Errorf("Expected stats field '%s' not found", field)
		}
	}

	// Validate specific values
	if stats["address"] != cfg.Server.Address {
		t.Errorf("Expected address %s, got %v", cfg.Server.Address, stats["address"])
	}

	if stats["port"] != cfg.Server.Port {
		t.Errorf("Expected port %d, got %v", cfg.Server.Port, stats["port"])
	}

	if stats["local_hosts"] != len(cfg.Hosts) {
		t.Errorf("Expected local_hosts %d, got %v", len(cfg.Hosts), stats["local_hosts"])
	}

	if stats["ttl"] != cfg.DNS.TTL {
		t.Errorf("Expected TTL %d, got %v", cfg.DNS.TTL, stats["ttl"])
	}

	if stats["recursion"] != cfg.DNS.EnableRecursion {
		t.Errorf("Expected recursion %v, got %v", cfg.DNS.EnableRecursion, stats["recursion"])
	}
}

func TestServerStartStop(t *testing.T) {
	// Use a unique port to avoid conflicts
	port := findFreePort()

	cfg := &config.Config{
		Server: config.ServerConfig{
			Address: "127.0.0.1",
			Port:    port,
		},
		DNS: config.DNSConfig{
			TTL:             300,
			EnableRecursion: false, // Disable for testing
		},
		HomeDNSDomain: "home",
	}

	server := NewServer(cfg)

	// Test starting the server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Start()
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Test that the server is listening
	conn, err := net.Dial("udp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Errorf("Could not connect to server: %v", err)
	} else {
		conn.Close()
	}

	// Stop the server
	err = server.Stop()
	if err != nil {
		t.Errorf("Error stopping server: %v", err)
	}

	// Check that Start() returned without error
	select {
	case err := <-errChan:
		if err != nil {
			t.Errorf("Server Start() returned error: %v", err)
		}
	case <-time.After(1 * time.Second):
		// Server should have stopped by now
	}
}

func TestServerStartPortInUse(t *testing.T) {
	port := findFreePort()

	// Bind to the port first
	listener, err := net.ListenPacket("udp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("Could not bind to test port: %v", err)
	}
	defer listener.Close()

	cfg := &config.Config{
		Server: config.ServerConfig{
			Address: "127.0.0.1",
			Port:    port,
		},
		DNS: config.DNSConfig{
			TTL:             300,
			EnableRecursion: false,
		},
		HomeDNSDomain: "home",
	}

	server := NewServer(cfg)

	// Should fail to start because port is in use
	err = server.Start()
	if err == nil {
		t.Error("Expected error when starting server on occupied port")
		server.Stop() // Clean up if it somehow started
	}
}

// Helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && s[len(s)-len(substr):] == substr ||
		(len(s) > len(substr) && findInString(s, substr))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func findFreePort() int {
	// Find a free port by binding to port 0
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		return 8053 // fallback
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return 8053 // fallback
	}
	defer conn.Close()

	return conn.LocalAddr().(*net.UDPAddr).Port
}

func BenchmarkServerGetStats(b *testing.B) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Address: "127.0.0.1",
			Port:    8053,
		},
		DNS: config.DNSConfig{
			TTL:             300,
			EnableRecursion: true,
		},
		Hosts: make([]config.HostEntry, 100), // Many hosts for benchmarking
		Upstream: []config.UpstreamServer{
			{Address: "8.8.8.8", Protocol: "udp", Port: 53},
		},
		HomeDNSDomain: "home",
	}

	// Fill hosts
	for i := 0; i < 100; i++ {
		cfg.Hosts[i] = config.HostEntry{
			Hostname: fmt.Sprintf("host%d.home", i),
			IPv4:     []string{fmt.Sprintf("192.168.1.%d", i%255)},
		}
	}

	server := NewServer(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server.GetStats()
	}
}
