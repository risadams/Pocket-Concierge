package dns

import (
	"fmt"
	"testing"

	"github.com/miekg/dns"
	"github.com/risadams/Pocket-Concierge/internal/config"
)

func TestNewHostCache(t *testing.T) {
	cfg := &config.Config{
		DNS: config.DNSConfig{
			TTL: 300,
		},
		HomeDNSDomain: "home",
		Hosts: []config.HostEntry{
			{
				Hostname: "test1",
				IPv4:     []string{"192.168.1.1"},
			},
		},
	}

	cache := NewHostCache(cfg)
	if cache == nil {
		t.Fatal("NewHostCache returned nil")
	}
	if cache.hosts == nil {
		t.Fatal("Cache hosts map not initialized")
	}
	if cache.records == nil {
		t.Fatal("Cache records map not initialized")
	}
}

func TestHostCache_Lookup(t *testing.T) {
	cfg := &config.Config{
		DNS: config.DNSConfig{
			TTL: 300,
		},
		HomeDNSDomain: "home",
		Hosts: []config.HostEntry{
			{
				Hostname: "test1",
				IPv4:     []string{"192.168.1.1", "192.168.1.2"},
				IPv6:     []string{"2001:db8::1"},
			},
		},
	}

	cache := NewHostCache(cfg)

	// Test lookup for non-existent entry
	entry, found := cache.Lookup("nonexistent.example.com")
	if found {
		t.Error("Expected not found for non-existent entry")
	}
	if entry != nil {
		t.Error("Expected nil entry for non-existent hostname")
	}

	// Test lookup for existing entry
	entry, found = cache.Lookup("test1")
	if !found {
		t.Error("Expected to find existing entry")
	}
	if entry == nil {
		t.Fatal("Expected non-nil entry")
	}
	if entry.Hostname != "test1" {
		t.Errorf("Expected hostname 'test1', got %s", entry.Hostname)
	}

	// Test lookup with domain
	entry, found = cache.Lookup("test1.home")
	if !found {
		t.Error("Expected to find entry with full domain")
	}

	// Test case insensitive lookup
	entry, found = cache.Lookup("TEST1")
	if !found {
		t.Error("Expected case insensitive lookup to work")
	}
}

func TestHostCache_LookupRecords(t *testing.T) {
	cfg := &config.Config{
		DNS: config.DNSConfig{
			TTL: 300,
		},
		HomeDNSDomain: "home",
		Hosts: []config.HostEntry{
			{
				Hostname: "test1",
				IPv4:     []string{"192.168.1.1", "192.168.1.2"},
				IPv6:     []string{"2001:db8::1"},
			},
		},
	}

	cache := NewHostCache(cfg)

	// Test A record lookup
	records := cache.LookupRecords("test1", dns.TypeA)
	if len(records) != 2 {
		t.Errorf("Expected 2 A records, got %d", len(records))
	}

	// Test AAAA record lookup
	records = cache.LookupRecords("test1", dns.TypeAAAA)
	if len(records) != 1 {
		t.Errorf("Expected 1 AAAA record, got %d", len(records))
	}

	// Test with full domain
	records = cache.LookupRecords("test1.home", dns.TypeA)
	if len(records) != 2 {
		t.Errorf("Expected 2 A records for full domain, got %d", len(records))
	}

	// Test unsupported record type
	records = cache.LookupRecords("test1", dns.TypeMX)
	if len(records) != 0 {
		t.Errorf("Expected 0 records for unsupported type, got %d", len(records))
	}

	// Test non-existent hostname
	records = cache.LookupRecords("nonexistent", dns.TypeA)
	if len(records) != 0 {
		t.Errorf("Expected 0 records for non-existent hostname, got %d", len(records))
	}
}

func TestHostCache_Rebuild(t *testing.T) {
	// Start with empty config
	cfg := &config.Config{
		DNS: config.DNSConfig{
			TTL: 300,
		},
		HomeDNSDomain: "home",
		Hosts:         []config.HostEntry{},
	}

	cache := NewHostCache(cfg)

	// Should be empty initially
	_, found := cache.Lookup("test1")
	if found {
		t.Error("Expected empty cache initially")
	}

	// Rebuild with new config
	cfg.Hosts = []config.HostEntry{
		{
			Hostname: "test1",
			IPv4:     []string{"192.168.1.1"},
		},
		{
			Hostname: "test2.example.com",
			IPv4:     []string{"10.0.0.1"},
		},
	}

	cache.Rebuild(cfg)

	// Should find new entries
	entry, found := cache.Lookup("test1")
	if !found {
		t.Error("Expected to find test1 after rebuild")
	}
	if entry.Hostname != "test1" {
		t.Errorf("Expected hostname 'test1', got %s", entry.Hostname)
	}

	entry, found = cache.Lookup("test2.example.com")
	if !found {
		t.Error("Expected to find test2.example.com after rebuild")
	}

	// Verify records were created
	records := cache.LookupRecords("test1", dns.TypeA)
	if len(records) != 1 {
		t.Errorf("Expected 1 A record for test1, got %d", len(records))
	}
}

func TestHostCache_DomainHandling(t *testing.T) {
	cfg := &config.Config{
		DNS: config.DNSConfig{
			TTL: 300,
		},
		HomeDNSDomain: "local",
		Hosts: []config.HostEntry{
			{
				Hostname: "simple",
				IPv4:     []string{"192.168.1.1"},
			},
			{
				Hostname: "full.domain.com",
				IPv4:     []string{"192.168.1.2"},
			},
		},
	}

	cache := NewHostCache(cfg)

	// Simple hostname should get domain added
	_, found := cache.Lookup("simple.local")
	if !found {
		t.Error("Expected to find simple hostname with added domain")
	}

	// Full hostname should be accessible as-is
	_, found = cache.Lookup("full.domain.com")
	if !found {
		t.Error("Expected to find full hostname as-is")
	}

	// Should also work with trailing dot
	_, found = cache.Lookup("simple.")
	if !found {
		t.Error("Expected to find hostname with trailing dot")
	}
}

func TestHostCache_RecordTTL(t *testing.T) {
	cfg := &config.Config{
		DNS: config.DNSConfig{
			TTL: 600,
		},
		HomeDNSDomain: "home",
		Hosts: []config.HostEntry{
			{
				Hostname: "test1",
				IPv4:     []string{"192.168.1.1"},
			},
		},
	}

	cache := NewHostCache(cfg)
	records := cache.LookupRecords("test1", dns.TypeA)

	if len(records) == 0 {
		t.Fatal("Expected at least one record")
	}

	if records[0].Header().Ttl != 600 {
		t.Errorf("Expected TTL 600, got %d", records[0].Header().Ttl)
	}
}

// Benchmark tests
func BenchmarkHostCache_Lookup(b *testing.B) {
	cfg := &config.Config{
		DNS: config.DNSConfig{
			TTL: 300,
		},
		HomeDNSDomain: "home",
		Hosts: []config.HostEntry{
			{
				Hostname: "test1",
				IPv4:     []string{"192.168.1.1", "192.168.1.2"},
			},
		},
	}

	cache := NewHostCache(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Lookup("test1")
	}
}

func BenchmarkHostCache_LookupRecords(b *testing.B) {
	cfg := &config.Config{
		DNS: config.DNSConfig{
			TTL: 300,
		},
		HomeDNSDomain: "home",
		Hosts: []config.HostEntry{
			{
				Hostname: "test1",
				IPv4:     []string{"192.168.1.1", "192.168.1.2"},
			},
		},
	}

	cache := NewHostCache(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.LookupRecords("test1", dns.TypeA)
	}
}

func BenchmarkHostCache_Rebuild(b *testing.B) {
	cfg := &config.Config{
		DNS: config.DNSConfig{
			TTL: 300,
		},
		HomeDNSDomain: "home",
		Hosts:         make([]config.HostEntry, 100),
	}

	for i := 0; i < 100; i++ {
		cfg.Hosts[i] = config.HostEntry{
			Hostname: fmt.Sprintf("host%d", i),
			IPv4:     []string{"192.168.1.1"},
		}
	}

	cache := NewHostCache(cfg)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Rebuild(cfg)
	}
}
