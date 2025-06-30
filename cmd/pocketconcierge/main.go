package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/risadams/Pocket-Concierge/internal/config"
	"github.com/risadams/Pocket-Concierge/internal/server"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}

func main() {
	fmt.Println("🏨 PocketConcierge DNS Server v0.1.0")
	fmt.Println("Starting your home network concierge...")

	// Load configuration
	configFile := "config.yaml"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		// log.printf("⚠️  Config loading failed: %v", err)
		log.Println("Using default configuration...")
		cfg = config.DefaultConfig()
	} else {
		fmt.Printf("✅ Loaded configuration from %s", configFile)
	}

	// Display configuration summary
	fmt.Printf("📋 Server: %s:%d\n", cfg.Server.Address, cfg.Server.Port)

	// Pretty print upstream servers
	fmt.Printf("📋 Upstream DNS servers: %d configured\n", len(cfg.Upstream))
	for i, upstream := range cfg.Upstream {
		name := upstream.Name
		if name == "" {
			name = fmt.Sprintf("Server %d", i+1)
		}
		if upstream.Port != 0 {
			fmt.Printf("   • %s: %s://%s:%d%s\n", name, upstream.Protocol, upstream.Address, upstream.Port, upstream.Path)
		} else {
			fmt.Printf("   • %s: %s://%s%s\n", name, upstream.Protocol, upstream.Address, upstream.Path)
		}
	}

	fmt.Printf("📋 Local hosts: %d configured\n", len(cfg.Hosts))
	if len(cfg.Hosts) > 0 {
		for _, host := range cfg.Hosts {
			fmt.Printf("   • %s", host.Hostname)
			if len(host.IPv4) > 0 {
				fmt.Printf(" → %v", host.IPv4)
			}
			if len(host.IPv6) > 0 {
				fmt.Printf(" → %v", host.IPv6)
			}
			fmt.Println()
		}
	}

	fmt.Printf("📋 Blocked domains: %d configured\n", len(cfg.DNS.BlockList))
	if len(cfg.DNS.BlockList) > 0 && len(cfg.DNS.BlockList) <= 5 {
		for _, domain := range cfg.DNS.BlockList {
			fmt.Printf("   • %s\n", domain)
		}
	} else if len(cfg.DNS.BlockList) > 5 {
		for i := 0; i < 3; i++ {
			fmt.Printf("   • %s\n", cfg.DNS.BlockList[i])
		}
		fmt.Printf("   • ... and %d more\n", len(cfg.DNS.BlockList)-3)
	}

	fmt.Printf("📋 Log level: %s\n", cfg.LogLevel)

	// Create and start server
	dnsServer := server.NewServer(cfg)

	// Graceful shutdown handling
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\n👋 PocketConcierge shutting down gracefully...")
		if err := dnsServer.Stop(); err != nil {
			fmt.Printf("❌ Error stopping server: %v", err)
		}
		os.Exit(0)
	}()

	// Start the DNS server
	log.Println("✅ Ready to serve your home network!")
	if err := dnsServer.Start(); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
