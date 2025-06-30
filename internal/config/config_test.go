package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		configYAML  string
		expectError bool
		validate    func(*testing.T, *Config)
	}{
		{
			name: "valid configuration",
			configYAML: `
server:
  port: 8053
  address: "127.0.0.1"
dns:
  ttl: 300
  enable_recursion: true
  cache_size: 1000
hosts:
  - hostname: "test.home"
    ipv4: ["192.168.1.100"]
    ipv6: ["2001:db8::1"]
upstream:
  - name: "cloudflare"
    address: "1.1.1.1"
    protocol: "udp"
    port: 53
    verify: true
log_level: "info"
home_dns_domain: "home"
`,
			expectError: false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Server.Port != 8053 {
					t.Errorf("Expected port 8053, got %d", cfg.Server.Port)
				}
				if cfg.Server.Address != "127.0.0.1" {
					t.Errorf("Expected address 127.0.0.1, got %s", cfg.Server.Address)
				}
				if cfg.DNS.TTL != 300 {
					t.Errorf("Expected TTL 300, got %d", cfg.DNS.TTL)
				}
				if !cfg.DNS.EnableRecursion {
					t.Error("Expected recursion enabled")
				}
				if len(cfg.Hosts) != 1 {
					t.Errorf("Expected 1 host, got %d", len(cfg.Hosts))
				}
				if cfg.Hosts[0].Hostname != "test.home" {
					t.Errorf("Expected hostname test.home, got %s", cfg.Hosts[0].Hostname)
				}
				if len(cfg.Upstream) != 1 {
					t.Errorf("Expected 1 upstream server, got %d", len(cfg.Upstream))
				}
				if cfg.LogLevel != "info" {
					t.Errorf("Expected log level info, got %s", cfg.LogLevel)
				}
			},
		},
		{
			name: "invalid YAML",
			configYAML: `
server:
  port: invalid
`,
			expectError: true,
		},
		{
			name: "missing required fields",
			configYAML: `
server:
  port: 8053
`,
			expectError: false, // Should use defaults
			validate: func(t *testing.T, cfg *Config) {
				if cfg.LogLevel == "" {
					t.Error("Expected default log level to be set")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")

			if err := os.WriteFile(configFile, []byte(tt.configYAML), 0644); err != nil {
				t.Fatalf("Failed to create temp config file: %v", err)
			}

			cfg, err := LoadConfig(configFile)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := LoadConfig("nonexistent.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	// Validate default values
	if cfg.Server.Port != 53 {
		t.Errorf("Expected default port 53, got %d", cfg.Server.Port)
	}

	if cfg.Server.Address != "0.0.0.0" {
		t.Errorf("Expected default address 0.0.0.0, got %s", cfg.Server.Address)
	}

	if cfg.DNS.TTL != 300 {
		t.Errorf("Expected default TTL 300, got %d", cfg.DNS.TTL)
	}

	if !cfg.DNS.EnableRecursion {
		t.Error("Expected default recursion enabled")
	}

	if cfg.LogLevel != "info" {
		t.Errorf("Expected default log level info, got %s", cfg.LogLevel)
	}

	if cfg.HomeDNSDomain != "home" {
		t.Errorf("Expected default home domain 'home', got %s", cfg.HomeDNSDomain)
	}

	// Should have at least one upstream server
	if len(cfg.Upstream) == 0 {
		t.Error("Expected at least one default upstream server")
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &Config{
				Server: ServerConfig{
					Port:    8053,
					Address: "127.0.0.1",
				},
				DNS: DNSConfig{
					TTL:             300,
					EnableRecursion: true,
					CacheSize:       1000,
				},
				Hosts: []HostEntry{
					{
						Hostname: "test.home",
						IPv4:     []string{"192.168.1.100"},
					},
				},
				Upstream: []UpstreamServer{
					{
						Address:  "1.1.1.1",
						Protocol: "udp",
						Port:     53,
					},
				},
				LogLevel:      "info",
				HomeDNSDomain: "home",
			},
			expectError: false,
		},
		{
			name: "invalid port - too low",
			config: &Config{
				Server: ServerConfig{
					Port:    0,
					Address: "127.0.0.1",
				},
				LogLevel:      "info",
				HomeDNSDomain: "home",
			},
			expectError: true,
			errorMsg:    "invalid port",
		},
		{
			name: "invalid port - too high",
			config: &Config{
				Server: ServerConfig{
					Port:    65536,
					Address: "127.0.0.1",
				},
				LogLevel:      "info",
				HomeDNSDomain: "home",
			},
			expectError: true,
			errorMsg:    "invalid port",
		},
		{
			name: "invalid server address",
			config: &Config{
				Server: ServerConfig{
					Port:    8053,
					Address: "invalid-ip",
				},
				LogLevel:      "info",
				HomeDNSDomain: "home",
			},
			expectError: true,
			errorMsg:    "invalid server address",
		},
		{
			name: "empty upstream address",
			config: &Config{
				Server: ServerConfig{
					Port:    8053,
					Address: "127.0.0.1",
				},
				Upstream: []UpstreamServer{
					{
						Address:  "",
						Protocol: "udp",
					},
				},
				LogLevel:      "info",
				HomeDNSDomain: "home",
			},
			expectError: true,
			errorMsg:    "address cannot be empty",
		},
		{
			name: "invalid upstream protocol",
			config: &Config{
				Server: ServerConfig{
					Port:    8053,
					Address: "127.0.0.1",
				},
				Upstream: []UpstreamServer{
					{
						Address:  "1.1.1.1",
						Protocol: "invalid",
					},
				},
				LogLevel:      "info",
				HomeDNSDomain: "home",
			},
			expectError: true,
			errorMsg:    "invalid protocol",
		},
		{
			name: "empty hostname",
			config: &Config{
				Server: ServerConfig{
					Port:    8053,
					Address: "127.0.0.1",
				},
				Hosts: []HostEntry{
					{
						Hostname: "",
						IPv4:     []string{"192.168.1.100"},
					},
				},
				LogLevel:      "info",
				HomeDNSDomain: "home",
			},
			expectError: true,
			errorMsg:    "hostname cannot be empty",
		},
		{
			name: "invalid IPv4 address",
			config: &Config{
				Server: ServerConfig{
					Port:    8053,
					Address: "127.0.0.1",
				},
				Hosts: []HostEntry{
					{
						Hostname: "test.home",
						IPv4:     []string{"invalid-ip"},
					},
				},
				LogLevel:      "info",
				HomeDNSDomain: "home",
			},
			expectError: true,
			errorMsg:    "invalid IPv4 address",
		},
		{
			name: "invalid IPv6 address",
			config: &Config{
				Server: ServerConfig{
					Port:    8053,
					Address: "127.0.0.1",
				},
				Hosts: []HostEntry{
					{
						Hostname: "test.home",
						IPv6:     []string{"invalid-ipv6"},
					},
				},
				LogLevel:      "info",
				HomeDNSDomain: "home",
			},
			expectError: true,
			errorMsg:    "invalid IPv6 address",
		},
		{
			name: "host without IP addresses",
			config: &Config{
				Server: ServerConfig{
					Port:    8053,
					Address: "127.0.0.1",
				},
				Hosts: []HostEntry{
					{
						Hostname: "test.home",
					},
				},
				LogLevel:      "info",
				HomeDNSDomain: "home",
			},
			expectError: true,
			errorMsg:    "must have at least one IP address",
		},
		{
			name: "invalid log level",
			config: &Config{
				Server: ServerConfig{
					Port:    8053,
					Address: "127.0.0.1",
				},
				LogLevel:      "invalid",
				HomeDNSDomain: "home",
			},
			expectError: true,
			errorMsg:    "invalid log level",
		},
		{
			name: "empty home DNS domain",
			config: &Config{
				Server: ServerConfig{
					Port:    8053,
					Address: "127.0.0.1",
				},
				LogLevel:      "info",
				HomeDNSDomain: "",
			},
			expectError: true,
			errorMsg:    "home_dns_domain cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				if err == nil {
					t.Error("Expected validation error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestConfigValidateUpstreamDefaults(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Port:    8053,
			Address: "127.0.0.1",
		},
		Upstream: []UpstreamServer{
			{
				Address:  "1.1.1.1",
				Protocol: "udp",
			},
			{
				Address:  "8.8.8.8",
				Protocol: "tcp",
			},
			{
				Address:  "9.9.9.9",
				Protocol: "tls",
			},
			{
				Address:  "cloudflare-dns.com",
				Protocol: "https",
			},
		},
		LogLevel:      "info",
		HomeDNSDomain: "home",
	}

	err := config.Validate()
	if err != nil {
		t.Errorf("Unexpected validation error: %v", err)
	}

	// Check that default ports were set
	expectedPorts := []int{53, 53, 853, 443}
	for i, expected := range expectedPorts {
		if config.Upstream[i].Port != expected {
			t.Errorf("Expected upstream %d port %d, got %d", i, expected, config.Upstream[i].Port)
		}
	}

	// Check HTTPS path default
	if config.Upstream[3].Path != "/dns-query" {
		t.Errorf("Expected HTTPS path '/dns-query', got '%s'", config.Upstream[3].Path)
	}
}

func TestConfigIsBlocked(t *testing.T) {
	cfg := &Config{
		DNS: DNSConfig{
			BlockList: []string{
				"blocked.example.com",
				"evil.net",
				"ads.tracker.org",
			},
		},
	}

	tests := []struct {
		name     string
		domain   string
		expected bool
	}{
		{
			name:     "exact match blocked",
			domain:   "blocked.example.com",
			expected: true,
		},
		{
			name:     "exact match blocked with trailing dot",
			domain:   "blocked.example.com.",
			expected: true,
		},
		{
			name:     "subdomain of blocked domain",
			domain:   "sub.evil.net",
			expected: true,
		},
		{
			name:     "deep subdomain of blocked domain",
			domain:   "deep.sub.evil.net",
			expected: true,
		},
		{
			name:     "subdomain with trailing dot",
			domain:   "sub.ads.tracker.org.",
			expected: true,
		},
		{
			name:     "not blocked domain",
			domain:   "good.example.com",
			expected: false,
		},
		{
			name:     "partial match should not block",
			domain:   "notevil.net",
			expected: false,
		},
		{
			name:     "similar but not subdomain",
			domain:   "evil.network",
			expected: false,
		},
		{
			name:     "empty domain",
			domain:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cfg.IsBlocked(tt.domain)
			if result != tt.expected {
				t.Errorf("IsBlocked(%q) = %v, expected %v", tt.domain, result, tt.expected)
			}
		})
	}
}

func BenchmarkLoadConfig(b *testing.B) {
	configYAML := `
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
`

	tmpDir := b.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte(configYAML), 0644); err != nil {
		b.Fatalf("Failed to create temp config file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := LoadConfig(configFile)
		if err != nil {
			b.Fatalf("LoadConfig failed: %v", err)
		}
	}
}

func BenchmarkConfigValidate(b *testing.B) {
	cfg := DefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := cfg.Validate()
		if err != nil {
			b.Fatalf("Validate failed: %v", err)
		}
	}
}
