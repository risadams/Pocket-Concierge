package dns

import (
	"fmt"
	"net"
	"testing"

	"github.com/miekg/dns"
	"github.com/risadams/Pocket-Concierge/internal/config"
)

func TestNewHandler(t *testing.T) {
	cfg := &config.Config{
		HomeDNSDomain: "home",
		DNS: config.DNSConfig{
			TTL:             300,
			EnableRecursion: true,
		},
		Upstream: []config.UpstreamServer{
			{
				Address:  "8.8.8.8",
				Protocol: "udp",
				Port:     53,
			},
		},
	}

	handler := NewHandler(cfg)
	if handler == nil {
		t.Fatal("NewHandler returned nil")
	}

	if handler.config != cfg {
		t.Error("Handler config not set correctly")
	}

	if handler.resolver == nil {
		t.Error("Resolver not initialized")
	}

	if handler.client == nil {
		t.Error("DNS client not initialized")
	}

	if handler.secureClient == nil {
		t.Error("Secure client not initialized")
	}
}

// MockResponseWriter implements dns.ResponseWriter for testing
type MockResponseWriter struct {
	responses  []*dns.Msg
	localAddr  net.Addr
	remoteAddr net.Addr
}

func (m *MockResponseWriter) LocalAddr() net.Addr {
	if m.localAddr == nil {
		return &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8053}
	}
	return m.localAddr
}

func (m *MockResponseWriter) RemoteAddr() net.Addr {
	if m.remoteAddr == nil {
		return &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}
	}
	return m.remoteAddr
}

func (m *MockResponseWriter) WriteMsg(msg *dns.Msg) error {
	m.responses = append(m.responses, msg)
	return nil
}

func (m *MockResponseWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

func (m *MockResponseWriter) Close() error {
	return nil
}

func (m *MockResponseWriter) TsigStatus() error {
	return nil
}

func (m *MockResponseWriter) TsigTimersOnly(bool) {}

func (m *MockResponseWriter) Hijack() {}

func TestHandlerServeDNSLocalResolution(t *testing.T) {
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
			TTL:             300,
			EnableRecursion: false, // Disable recursion for local-only test
		},
	}

	handler := NewHandler(cfg)
	writer := &MockResponseWriter{}

	tests := []struct {
		name          string
		hostname      string
		qtype         uint16
		expectedCount int
	}{
		{
			name:          "A record query",
			hostname:      "test.home.",
			qtype:         dns.TypeA,
			expectedCount: 4, // 2 IPv4 addresses x 2 (short+full name) = 4 records
		},
		{
			name:          "AAAA record query",
			hostname:      "test.home.",
			qtype:         dns.TypeAAAA,
			expectedCount: 2, // 1 IPv6 address x 2 (short+full name) = 2 records
		},
		{
			name:          "nonexistent host",
			hostname:      "nonexistent.home.",
			qtype:         dns.TypeA,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create DNS query
			query := &dns.Msg{}
			query.SetQuestion(tt.hostname, tt.qtype)
			query.Id = 12345

			// Reset writer
			writer.responses = nil

			// Handle the query
			handler.ServeDNS(writer, query)

			// Verify response
			if len(writer.responses) != 1 {
				t.Fatalf("Expected 1 response, got %d", len(writer.responses))
			}

			response := writer.responses[0]
			if response.Id != query.Id {
				t.Errorf("Expected response ID %d, got %d", query.Id, response.Id)
			}

			if !response.Authoritative {
				t.Error("Expected authoritative response")
			}

			if len(response.Answer) != tt.expectedCount {
				t.Errorf("Expected %d answers, got %d", tt.expectedCount, len(response.Answer))
			}

			// Validate answer records
			for _, answer := range response.Answer {
				if answer.Header().Rrtype != tt.qtype {
					t.Errorf("Expected answer type %d, got %d", tt.qtype, answer.Header().Rrtype)
				}
			}
		})
	}
}

func TestHandlerServeDNSMultipleQuestions(t *testing.T) {
	cfg := &config.Config{
		HomeDNSDomain: "home",
		Hosts: []config.HostEntry{
			{
				Hostname: "test1.home",
				IPv4:     []string{"192.168.1.100"},
			},
			{
				Hostname: "test2.home",
				IPv4:     []string{"192.168.1.101"},
			},
		},
		DNS: config.DNSConfig{
			TTL:             300,
			EnableRecursion: false,
		},
	}

	handler := NewHandler(cfg)
	writer := &MockResponseWriter{}

	// Create query with multiple questions
	query := &dns.Msg{}
	query.Id = 12345
	query.Question = []dns.Question{
		{Name: "test1.home.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
		{Name: "test2.home.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
	}

	handler.ServeDNS(writer, query)

	// Verify response
	if len(writer.responses) != 1 {
		t.Fatalf("Expected 1 response, got %d", len(writer.responses))
	}

	response := writer.responses[0]
	if len(response.Answer) != 4 {
		t.Errorf("Expected 4 answers for 2 questions, got %d", len(response.Answer))
	}
}

func TestHandlerServeDNSEmptyResponse(t *testing.T) {
	cfg := &config.Config{
		HomeDNSDomain: "home",
		DNS: config.DNSConfig{
			TTL:             300,
			EnableRecursion: false,
		},
		Hosts: []config.HostEntry{}, // No hosts configured
	}

	handler := NewHandler(cfg)
	writer := &MockResponseWriter{}

	// Create query for nonexistent host
	query := &dns.Msg{}
	query.SetQuestion("nonexistent.home.", dns.TypeA)
	query.Id = 12345

	handler.ServeDNS(writer, query)

	// Verify response
	if len(writer.responses) != 1 {
		t.Fatalf("Expected 1 response, got %d", len(writer.responses))
	}

	response := writer.responses[0]
	if len(response.Answer) != 0 {
		t.Errorf("Expected 0 answers for nonexistent host, got %d", len(response.Answer))
	}

	if !response.Authoritative {
		t.Error("Expected authoritative response even when no answers")
	}
}

// MockSecureClient for testing upstream forwarding
type MockSecureClient struct {
	responses map[string]*dns.Msg
	errors    map[string]error
}

func (m *MockSecureClient) Query(query *dns.Msg, upstream config.UpstreamServer) (*dns.Msg, error) {
	key := query.Question[0].Name
	if err, exists := m.errors[key]; exists {
		return nil, err
	}
	if response, exists := m.responses[key]; exists {
		return response, nil
	}
	return nil, fmt.Errorf("timeout")
}

func TestHandlerForwardUpstream(t *testing.T) {
	// Note: This test focuses on the forwardUpstream method behavior
	// In a real scenario, this would test the actual upstream forwarding
	cfg := &config.Config{
		HomeDNSDomain: "home",
		DNS: config.DNSConfig{
			TTL:             300,
			EnableRecursion: true,
		},
		Upstream: []config.UpstreamServer{
			{
				Address:  "8.8.8.8",
				Protocol: "udp",
				Port:     53,
			},
		},
		Hosts: []config.HostEntry{}, // No local hosts
	}

	handler := NewHandler(cfg)

	// Create a query for external domain
	query := &dns.Msg{}
	query.SetQuestion("google.com.", dns.TypeA)
	question := query.Question[0]

	// Test forwardUpstream method
	// Since we can't mock the actual upstream without more complex setup,
	// we test that the method doesn't panic and returns expected structure
	answers := handler.forwardUpstream(question, query)

	// The method should return a slice (empty or with answers)
	// In real scenarios, this would contain upstream responses
	if answers == nil {
		t.Error("forwardUpstream should return a slice, even if empty")
	}
}

func TestHandlerRecursionDisabled(t *testing.T) {
	cfg := &config.Config{
		HomeDNSDomain: "home",
		DNS: config.DNSConfig{
			TTL:             300,
			EnableRecursion: false, // Recursion disabled
		},
		Upstream: []config.UpstreamServer{
			{
				Address:  "8.8.8.8",
				Protocol: "udp",
				Port:     53,
			},
		},
		Hosts: []config.HostEntry{}, // No local hosts
	}

	handler := NewHandler(cfg)
	writer := &MockResponseWriter{}

	// Query for external domain (should not be forwarded)
	query := &dns.Msg{}
	query.SetQuestion("google.com.", dns.TypeA)
	query.Id = 12345

	handler.ServeDNS(writer, query)

	// Verify response has no answers (not forwarded)
	if len(writer.responses) != 1 {
		t.Fatalf("Expected 1 response, got %d", len(writer.responses))
	}

	response := writer.responses[0]
	if len(response.Answer) != 0 {
		t.Errorf("Expected 0 answers with recursion disabled, got %d", len(response.Answer))
	}
}

func BenchmarkHandlerServeDNS(b *testing.B) {
	cfg := &config.Config{
		HomeDNSDomain: "home",
		Hosts: []config.HostEntry{
			{
				Hostname: "test.home",
				IPv4:     []string{"192.168.1.100"},
			},
		},
		DNS: config.DNSConfig{
			TTL:             300,
			EnableRecursion: false,
		},
	}

	handler := NewHandler(cfg)
	writer := &MockResponseWriter{}

	query := &dns.Msg{}
	query.SetQuestion("test.home.", dns.TypeA)
	query.Id = 12345

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer.responses = nil // Reset for each iteration
		handler.ServeDNS(writer, query)
	}
}

func BenchmarkHandlerServeDNSConcurrent(b *testing.B) {
	cfg := &config.Config{
		HomeDNSDomain: "home",
		Hosts: []config.HostEntry{
			{
				Hostname: "test.home",
				IPv4:     []string{"192.168.1.100"},
			},
		},
		DNS: config.DNSConfig{
			TTL:             300,
			EnableRecursion: false,
		},
	}

	handler := NewHandler(cfg)

	query := &dns.Msg{}
	query.SetQuestion("test.home.", dns.TypeA)
	query.Id = 12345

	b.RunParallel(func(pb *testing.PB) {
		writer := &MockResponseWriter{}
		for pb.Next() {
			writer.responses = nil
			handler.ServeDNS(writer, query)
		}
	})
}
