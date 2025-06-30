package dns

import (
	"testing"

	"github.com/miekg/dns"
	"github.com/risadams/Pocket-Concierge/internal/config"
)

func TestNewResolver(t *testing.T) {
	cfg := &config.Config{
		HomeDNSDomain: "home",
		Hosts: []config.HostEntry{
			{
				Hostname: "test.home",
				IPv4:     []string{"192.168.1.100"},
			},
		},
		DNS: config.DNSConfig{
			TTL: 300,
		},
	}

	resolver := NewResolver(cfg)
	if resolver == nil {
		t.Fatal("NewResolver returned nil")
	}

	if resolver.config != cfg {
		t.Error("Resolver config not set correctly")
	}

	if resolver.hostCache == nil {
		t.Error("Host cache not initialized")
	}
}

func TestResolverResolveFast(t *testing.T) {
	cfg := &config.Config{
		HomeDNSDomain: "home",
		Hosts: []config.HostEntry{
			{
				Hostname: "test.home",
				IPv4:     []string{"192.168.1.100", "192.168.1.101"},
				IPv6:     []string{"2001:db8::1"},
			},
			{
				Hostname: "ipv6only.home",
				IPv6:     []string{"2001:db8::2"},
			},
		},
		DNS: config.DNSConfig{
			TTL: 300,
		},
	}

	resolver := NewResolver(cfg)

	tests := []struct {
		name          string
		hostname      string
		qtype         uint16
		expectedCount int
	}{
		{
			name:          "A records found",
			hostname:      "test.home.",
			qtype:         dns.TypeA,
			expectedCount: 4, // 2 IPv4 addresses x 2 (short+full name) = 4 records
		},
		{
			name:          "AAAA records found",
			hostname:      "test.home.",
			qtype:         dns.TypeAAAA,
			expectedCount: 2, // 1 IPv6 address x 2 (short+full name) = 2 records
		},
		{
			name:          "IPv6 only host",
			hostname:      "ipv6only.home.",
			qtype:         dns.TypeAAAA,
			expectedCount: 2, // 1 IPv6 address x 2 (short+full name) = 2 records
		},
		{
			name:          "IPv6 only host queried for A",
			hostname:      "ipv6only.home.",
			qtype:         dns.TypeA,
			expectedCount: 0,
		},
		{
			name:          "nonexistent host",
			hostname:      "nonexistent.home.",
			qtype:         dns.TypeA,
			expectedCount: 0,
		},
		{
			name:          "unsupported record type",
			hostname:      "test.home.",
			qtype:         dns.TypeMX,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			records := resolver.ResolveFast(tt.hostname, tt.qtype)

			if len(records) != tt.expectedCount {
				t.Errorf("Expected %d records, got %d", tt.expectedCount, len(records))
			}

			// Validate record types
			for _, record := range records {
				if record.Header().Rrtype != tt.qtype {
					t.Errorf("Expected record type %d, got %d", tt.qtype, record.Header().Rrtype)
				}
			}
		})
	}
}

func TestResolverResolveLocal(t *testing.T) {
	cfg := &config.Config{
		HomeDNSDomain: "home",
		Hosts: []config.HostEntry{
			{
				Hostname: "test.home",
				IPv4:     []string{"192.168.1.100"},
				IPv6:     []string{"2001:db8::1"},
			},
		},
		DNS: config.DNSConfig{
			TTL: 300,
		},
	}

	resolver := NewResolver(cfg)

	tests := []struct {
		name     string
		hostname string
		found    bool
	}{
		{
			name:     "existing host",
			hostname: "test.home",
			found:    true,
		},
		{
			name:     "case insensitive",
			hostname: "TEST.HOME",
			found:    true,
		},
		{
			name:     "nonexistent host",
			hostname: "nonexistent.home",
			found:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, found := resolver.ResolveLocal(tt.hostname)

			if found != tt.found {
				t.Errorf("Expected found=%v, got found=%v", tt.found, found)
			}

			if tt.found {
				if host == nil {
					t.Error("Expected host entry but got nil")
				} else {
					if len(host.IPv4) == 0 && len(host.IPv6) == 0 {
						t.Error("Expected host to have IP addresses")
					}
				}
			} else {
				if host != nil {
					t.Error("Expected nil host entry but got one")
				}
			}
		})
	}
}

func TestResolverGetAllHosts(t *testing.T) {
	hosts := []config.HostEntry{
		{
			Hostname: "test1.home",
			IPv4:     []string{"192.168.1.100"},
		},
		{
			Hostname: "test2.home",
			IPv4:     []string{"192.168.1.101"},
		},
	}

	cfg := &config.Config{
		HomeDNSDomain: "home",
		Hosts:         hosts,
		DNS: config.DNSConfig{
			TTL: 300,
		},
	}

	resolver := NewResolver(cfg)
	allHosts := resolver.GetAllHosts()

	if len(allHosts) != len(hosts) {
		t.Errorf("Expected %d hosts, got %d", len(hosts), len(allHosts))
	}

	// Verify hosts match
	for i, expectedHost := range hosts {
		if allHosts[i].Hostname != expectedHost.Hostname {
			t.Errorf("Host %d: expected hostname %s, got %s", i, expectedHost.Hostname, allHosts[i].Hostname)
		}
	}
}

func TestResolverAddHost(t *testing.T) {
	cfg := &config.Config{
		HomeDNSDomain: "home",
		Hosts: []config.HostEntry{
			{
				Hostname: "existing.home",
				IPv4:     []string{"192.168.1.100"},
			},
		},
		DNS: config.DNSConfig{
			TTL: 300,
		},
	}

	resolver := NewResolver(cfg)

	// Verify existing host
	if _, found := resolver.ResolveLocal("existing.home"); !found {
		t.Error("Expected to find existing host")
	}

	// Add new host
	newHost := config.HostEntry{
		Hostname: "new.home",
		IPv4:     []string{"192.168.1.200"},
	}
	resolver.AddHost(newHost)

	// Verify new host was added
	if _, found := resolver.ResolveLocal("new.home"); !found {
		t.Error("Expected to find newly added host")
	}

	// Verify host count increased
	allHosts := resolver.GetAllHosts()
	if len(allHosts) != 2 {
		t.Errorf("Expected 2 hosts after adding, got %d", len(allHosts))
	}

	// Verify DNS records were rebuilt
	records := resolver.ResolveFast("new.home.", dns.TypeA)
	if len(records) != 2 {
		t.Errorf("Expected 2 DNS record for new host, got %d", len(records))
	}
}

func TestResolverConcurrency(t *testing.T) {
	cfg := &config.Config{
		HomeDNSDomain: "home",
		Hosts: []config.HostEntry{
			{
				Hostname: "test.home",
				IPv4:     []string{"192.168.1.100"},
			},
		},
		DNS: config.DNSConfig{
			TTL: 300,
		},
	}

	resolver := NewResolver(cfg)

	// Test concurrent access
	const numGoroutines = 10
	const numQueries = 100

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < numQueries; j++ {
				// Mix different types of operations
				switch j % 3 {
				case 0:
					resolver.ResolveFast("test.home.", dns.TypeA)
				case 1:
					resolver.ResolveLocal("test.home")
				case 2:
					resolver.GetAllHosts()
				}
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify resolver still works after concurrent access
	records := resolver.ResolveFast("test.home.", dns.TypeA)
	if len(records) != 2 {
		t.Errorf("Expected 2 record after concurrent access, got %d", len(records))
	}
}

func BenchmarkResolverResolveFast(b *testing.B) {
	cfg := &config.Config{
		HomeDNSDomain: "home",
		Hosts: []config.HostEntry{
			{
				Hostname: "test.home",
				IPv4:     []string{"192.168.1.100", "192.168.1.101"},
				IPv6:     []string{"2001:db8::1"},
			},
		},
		DNS: config.DNSConfig{
			TTL: 300,
		},
	}

	resolver := NewResolver(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resolver.ResolveFast("test.home.", dns.TypeA)
	}
}

func BenchmarkResolverResolveLocal(b *testing.B) {
	cfg := &config.Config{
		HomeDNSDomain: "home",
		Hosts: []config.HostEntry{
			{
				Hostname: "test.home",
				IPv4:     []string{"192.168.1.100"},
			},
		},
		DNS: config.DNSConfig{
			TTL: 300,
		},
	}

	resolver := NewResolver(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resolver.ResolveLocal("test.home")
	}
}

func BenchmarkResolverConcurrentAccess(b *testing.B) {
	cfg := &config.Config{
		HomeDNSDomain: "home",
		Hosts: []config.HostEntry{
			{
				Hostname: "test.home",
				IPv4:     []string{"192.168.1.100"},
			},
		},
		DNS: config.DNSConfig{
			TTL: 300,
		},
	}

	resolver := NewResolver(cfg)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resolver.ResolveFast("test.home.", dns.TypeA)
		}
	})
}
