package server

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/miekg/dns"
	"github.com/risadams/Pocket-Concierge/internal/config"
	dnshandler "github.com/risadams/Pocket-Concierge/internal/dns"
)

// Server represents the PocketConcierge DNS server
type Server struct {
	config     *config.Config
	dnsHandler *dnshandler.Handler
	server     *dns.Server
}

// NewServer creates a new PocketConcierge server with optimized settings
func NewServer(cfg *config.Config) *Server {
	handler := dnshandler.NewHandler(cfg)

	server := &dns.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Address, cfg.Server.Port),
		Net:          "udp",
		Handler:      handler,
		ReadTimeout:  3 * time.Second, // Reduced timeout
		WriteTimeout: 3 * time.Second, // Reduced timeout
		UDPSize:      65535,           // Maximum UDP packet size
	}

	return &Server{
		config:     cfg,
		dnsHandler: handler,
		server:     server,
	}
}

// Start begins serving DNS requests
func (s *Server) Start() error {
	log.Printf("🚀 Starting DNS server on %s", s.server.Addr)

	// Check if we can bind to the port
	if err := s.checkPort(); err != nil {
		return fmt.Errorf("port check failed: %w", err)
	}

	// Start the server
	if err := s.server.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to start DNS server: %w", err)
	}

	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop() error {
	log.Println("🛑 Stopping DNS server...")
	return s.server.Shutdown()
}

// checkPort verifies we can bind to the configured port
func (s *Server) checkPort() error {
	addr := fmt.Sprintf("%s:%d", s.config.Server.Address, s.config.Server.Port)

	// Try to bind to the port
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		if s.config.Server.Port == 53 {
			return fmt.Errorf("cannot bind to port 53 (requires admin/root privileges). Try port 5353 for testing: %w", err)
		}
		return err
	}

	// Close immediately - we just wanted to test
	conn.Close()
	return nil
}

// GetStats returns basic server statistics
func (s *Server) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"address":      s.config.Server.Address,
		"port":         s.config.Server.Port,
		"upstream_dns": s.config.Upstream,
		"local_hosts":  len(s.config.Hosts),
		"ttl":          s.config.DNS.TTL,
		"recursion":    s.config.DNS.EnableRecursion,
	}
}
