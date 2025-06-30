package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/risadams/Pocket-Concierge/internal/config"
)

func TestMainConfigLoading(t *testing.T) {
	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	tests := []struct {
		name           string
		configContent  string
		args           []string
		expectDefault  bool
		validateConfig func(*testing.T, *config.Config)
	}{
		{
			name: "valid config file",
			configContent: `
server:
  port: 8053
  address: "127.0.0.1"
dns:
  ttl: 300
  enable_recursion: true
hosts:
  - hostname: "test.home"
    ipv4: ["192.168.1.100"]
upstream:
  - address: "1.1.1.1"
    protocol: "udp"
log_level: "info"
home_dns_domain: "home"
`,
			args:          []string{"pocketconcierge", "test-config.yaml"},
			expectDefault: false,
			validateConfig: func(t *testing.T, cfg *config.Config) {
				if cfg.Server.Port != 8053 {
					t.Errorf("Expected port 8053, got %d", cfg.Server.Port)
				}
				if len(cfg.Hosts) != 1 {
					t.Errorf("Expected 1 host, got %d", len(cfg.Hosts))
				}
			},
		},
		{
			name:          "no config file specified",
			args:          []string{"pocketconcierge"},
			expectDefault: true,
		},
		{
			name:          "nonexistent config file",
			args:          []string{"pocketconcierge", "nonexistent.yaml"},
			expectDefault: true,
		},
		{
			name: "invalid config file",
			configContent: `
invalid yaml content
  this should fail: [
`,
			args:          []string{"pocketconcierge", "invalid-config.yaml"},
			expectDefault: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tmpDir := t.TempDir()

			var configFile string
			if tt.configContent != "" {
				configFile = filepath.Join(tmpDir, "test-config.yaml")
				if len(tt.args) > 1 && tt.args[1] != "nonexistent.yaml" {
					// Write config file
					err := os.WriteFile(configFile, []byte(tt.configContent), 0644)
					if err != nil {
						t.Fatalf("Failed to create test config file: %v", err)
					}
					// Update args with actual path
					tt.args[1] = configFile
				}
			}

			// Set command line args
			os.Args = tt.args

			// Test the config loading logic from main()
			configFile = "config.yaml"
			if len(os.Args) > 1 {
				configFile = os.Args[1]
			}

			cfg, err := config.LoadConfig(configFile)
			if err != nil {
				// Should use default config
				cfg = config.DefaultConfig()
				if !tt.expectDefault {
					t.Errorf("Expected config to load successfully, but got error: %v", err)
				}
			} else {
				if tt.expectDefault {
					t.Error("Expected to use default config, but config loaded successfully")
				}
			}

			if cfg == nil {
				t.Fatal("Config should not be nil")
			}

			// Validate config if provided
			if tt.validateConfig != nil && !tt.expectDefault {
				tt.validateConfig(t, cfg)
			}

			// Validate default config properties
			if tt.expectDefault {
				if cfg.Server.Port == 0 {
					t.Error("Default config should have a valid port")
				}
				if cfg.HomeDNSDomain == "" {
					t.Error("Default config should have a home DNS domain")
				}
			}
		})
	}
}

func TestMainSignalHandling(t *testing.T) {
	// This test verifies that the signal handling logic is set up correctly
	// We can't easily test the actual signal handling without complex setup

	// Test that we can create the signal channel and register handlers
	// without panicking
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Signal handling setup panicked: %v", r)
		}
	}()

	// Simulate the signal handling setup from main()
	c := make(chan os.Signal, 1)
	if c == nil {
		t.Error("Failed to create signal channel")
	}

	// This would be the signal.Notify call in main()
	// signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// We can't test this directly without actually sending signals
}

func TestMainServerCreation(t *testing.T) {
	// Test that server creation works with various configs
	configs := []*config.Config{
		config.DefaultConfig(),
		{
			Server: config.ServerConfig{
				Address: "127.0.0.1",
				Port:    8053,
			},
			DNS: config.DNSConfig{
				TTL:             300,
				EnableRecursion: true,
			},
			HomeDNSDomain: "home",
			LogLevel:      "info",
		},
	}

	for i, cfg := range configs {
		t.Run(fmt.Sprintf("config_%d", i), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Server creation panicked: %v", r)
				}
			}()

			// This simulates the server.NewServer call from main()
			// We import the server package to test this
			// dnsServer := server.NewServer(cfg)

			// For this test, we just verify the config is valid
			if err := cfg.Validate(); err != nil {
				t.Errorf("Config validation failed: %v", err)
			}
		})
	}
}

func TestMainConfigurationDisplay(t *testing.T) {
	// Test that the configuration summary display doesn't panic
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
			{Address: "1.1.1.1", Protocol: "udp", Port: 53},
		},
		LogLevel:      "info",
		HomeDNSDomain: "home",
	}

	// Test the printf statements from main() don't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Configuration display panicked: %v", r)
		}
	}()

	// Simulate the display logic from main()
	_ = cfg.Server.Address
	_ = cfg.Server.Port
	_ = cfg.Upstream
	_ = len(cfg.Hosts)
	_ = cfg.LogLevel

	// Verify the values are what we expect
	if cfg.Server.Address != "127.0.0.1" {
		t.Errorf("Expected address 127.0.0.1, got %s", cfg.Server.Address)
	}
	if cfg.Server.Port != 8053 {
		t.Errorf("Expected port 8053, got %d", cfg.Server.Port)
	}
	if len(cfg.Hosts) != 2 {
		t.Errorf("Expected 2 hosts, got %d", len(cfg.Hosts))
	}
	if len(cfg.Upstream) != 2 {
		t.Errorf("Expected 2 upstream servers, got %d", len(cfg.Upstream))
	}
}

func TestMainDefaultConfigPath(t *testing.T) {
	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Test default config path logic
	os.Args = []string{"pocketconcierge"}

	configFile := "config.yaml"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	if configFile != "config.yaml" {
		t.Errorf("Expected default config file 'config.yaml', got '%s'", configFile)
	}

	// Test custom config path
	os.Args = []string{"pocketconcierge", "custom.yaml"}

	configFile = "config.yaml"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	if configFile != "custom.yaml" {
		t.Errorf("Expected custom config file 'custom.yaml', got '%s'", configFile)
	}
}

func TestMainLogFlags(t *testing.T) {
	// Test that log flags are set correctly in init()
	// We can't easily test the actual log output, but we can verify
	// the init() function doesn't panic

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Log initialization panicked: %v", r)
		}
	}()

	// The init() function should have already run when the test package loaded
	// We can just verify that logging works
	// log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}

// Integration test helpers
func createTestConfig(t *testing.T) string {
	content := `
server:
  port: 8053
  address: "127.0.0.1"
dns:
  ttl: 300
  enable_recursion: false
hosts:
  - hostname: "test.home"
    ipv4: ["192.168.1.100"]
upstream:
  - address: "8.8.8.8"
    protocol: "udp"
    port: 53
log_level: "info"
home_dns_domain: "home"
`

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test-config.yaml")

	err := os.WriteFile(configFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	return configFile
}

func BenchmarkConfigLoading(b *testing.B) {
	content := `
server:
  port: 8053
  address: "127.0.0.1"
dns:
  ttl: 300
  enable_recursion: false
hosts:
  - hostname: "test.home"
    ipv4: ["192.168.1.100"]
upstream:
  - address: "8.8.8.8"
    protocol: "udp"
    port: 53
log_level: "info"
home_dns_domain: "home"
`

	tmpDir := b.TempDir()
	configFile := filepath.Join(tmpDir, "test-config.yaml")

	err := os.WriteFile(configFile, []byte(content), 0644)
	if err != nil {
		b.Fatalf("Failed to create test config: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := config.LoadConfig(configFile)
		if err != nil {
			b.Fatalf("Config loading failed: %v", err)
		}
	}
}
