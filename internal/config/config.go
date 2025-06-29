package config

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the complete PocketConcierge configuration
type Config struct {
	Server   ServerConfig `yaml:"server"`
	DNS      DNSConfig    `yaml:"dns"`
	Hosts    []HostEntry  `yaml:"hosts"`
	Upstream []string     `yaml:"upstream"`
	LogLevel string       `yaml:"log_level"`
}

// ServerConfig defines server-specific settings
type ServerConfig struct {
	Port    int    `yaml:"port"`
	Address string `yaml:"address"`
}

// DNSConfig defines DNS-specific settings
type DNSConfig struct {
	TTL             int  `yaml:"ttl"`
	EnableRecursion bool `yaml:"enable_recursion"`
	CacheSize       int  `yaml:"cache_size"`
}

// HostEntry represents a hostname to IP mapping
type HostEntry struct {
	Hostname string   `yaml:"hostname"`
	IPv4     []string `yaml:"ipv4,omitempty"`
	IPv6     []string `yaml:"ipv6,omitempty"`
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:    53,
			Address: "0.0.0.0",
		},
		DNS: DNSConfig{
			TTL:             300, // 5 minutes
			EnableRecursion: true,
			CacheSize:       1000,
		},
		Upstream: []string{
			"8.8.8.8:53", // Google DNS
			"1.1.1.1:53", // Cloudflare DNS
		},
		LogLevel: "info",
		Hosts:    []HostEntry{},
	}
}

// LoadConfig reads configuration from a YAML file
func LoadConfig(filename string) (*Config, error) {
	// Start with defaults
	config := DefaultConfig()

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return config, fmt.Errorf("config file not found: %s", filename)
	}

	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// SaveConfig writes configuration to a YAML file
func (c *Config) SaveConfig(filename string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate server settings
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be 1-65535)", c.Server.Port)
	}

	if net.ParseIP(c.Server.Address) == nil && c.Server.Address != "0.0.0.0" {
		return fmt.Errorf("invalid server address: %s", c.Server.Address)
	}

	// Validate upstream servers
	for _, upstream := range c.Upstream {
		if _, _, err := net.SplitHostPort(upstream); err != nil {
			return fmt.Errorf("invalid upstream server format: %s", upstream)
		}
	}

	// Validate host entries
	for i, host := range c.Hosts {
		if host.Hostname == "" {
			return fmt.Errorf("host entry %d: hostname cannot be empty", i)
		}

		// Validate IPv4 addresses
		for _, ip := range host.IPv4 {
			if net.ParseIP(ip) == nil {
				return fmt.Errorf("host entry %s: invalid IPv4 address: %s", host.Hostname, ip)
			}
		}

		// Validate IPv6 addresses
		for _, ip := range host.IPv6 {
			if net.ParseIP(ip) == nil {
				return fmt.Errorf("host entry %s: invalid IPv6 address: %s", host.Hostname, ip)
			}
		}

		// Must have at least one IP
		if len(host.IPv4) == 0 && len(host.IPv6) == 0 {
			return fmt.Errorf("host entry %s: must have at least one IP address", host.Hostname)
		}
	}

	// Validate log level
	validLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true,
	}
	if !validLevels[c.LogLevel] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", c.LogLevel)
	}

	return nil
}

// GetHostByName returns the HostEntry for a given hostname
func (c *Config) GetHostByName(hostname string) (*HostEntry, bool) {
	for _, host := range c.Hosts {
		if host.Hostname == hostname {
			return &host, true
		}
	}
	return nil, false
}
