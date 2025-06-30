package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/miekg/dns"
	"github.com/risadams/Pocket-Concierge/internal/config"
)

func getDefaultServer() string {
	// Try to load config to get the port
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		// Try example config
		cfg, err = config.LoadConfig("configs/example.yaml")
		if err != nil {
			// Use default port
			return "127.0.0.1:8053"
		}
	}
	return fmt.Sprintf("127.0.0.1:%d", cfg.Server.Port)
}

func main() {
	if len(os.Args) < 2 {
		defaultServer := getDefaultServer()
		fmt.Println("Usage: go run test-dns.go <hostname> [server:port] [record-type]")
		fmt.Printf("Example: go run test-dns.go ris-laptop.home %s\n", defaultServer)
		fmt.Printf("         go run test-dns.go ris-laptop.home %s AAAA\n", defaultServer)
		fmt.Printf("         go run test-dns.go ris-laptop.home %s BOTH\n", defaultServer)
		os.Exit(1)
	}

	hostname := os.Args[1]
	server := getDefaultServer()
	recordType := "A"

	if len(os.Args) > 2 {
		server = os.Args[2]
	}
	if len(os.Args) > 3 {
		recordType = strings.ToUpper(os.Args[3])
	}

	client := &dns.Client{}

	if recordType == "BOTH" {
		// Query both A and AAAA records
		fmt.Printf("🔍 Querying %s for %s (A + AAAA records)\n", server, hostname)
		queryRecord(client, hostname, server, dns.TypeA, "A")
		fmt.Println()
		queryRecord(client, hostname, server, dns.TypeAAAA, "AAAA")
	} else {
		var qtype uint16
		switch recordType {
		case "A":
			qtype = dns.TypeA
		case "AAAA":
			qtype = dns.TypeAAAA
		case "ANY":
			qtype = dns.TypeANY
		default:
			log.Fatalf("❌ Unsupported record type: %s (use A, AAAA, ANY, or BOTH)", recordType)
		}

		fmt.Printf("🔍 Querying %s for %s (%s record)\n", server, hostname, recordType)
		queryRecord(client, hostname, server, qtype, recordType)
	}
}

func queryRecord(client *dns.Client, hostname, server string, qtype uint16, typeName string) {
	msg := &dns.Msg{}
	msg.SetQuestion(dns.Fqdn(hostname), qtype)

	response, _, err := client.Exchange(msg, server)
	if err != nil {
		fmt.Printf("❌ %s query failed: %v\n", typeName, err)
		return
	}

	if len(response.Answer) == 0 {
		fmt.Printf("❌ No %s answers received\n", typeName)
		return
	}

	fmt.Printf("✅ Received %d %s answer(s):\n", len(response.Answer), typeName)
	for _, answer := range response.Answer {
		fmt.Printf("   %s\n", answer.String())
	}
}
