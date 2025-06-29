package main

import (
	"fmt"
	"log"
	"os"

	"github.com/miekg/dns"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test-dns.go <hostname> [server:port]")
		fmt.Println("Example: go run test-dns.go ris-desktop.home 127.0.0.1:8053")
		os.Exit(1)
	}

	hostname := os.Args[1]
	server := "127.0.0.1:8053"
	if len(os.Args) > 2 {
		server = os.Args[2]
	}

	// Create DNS client
	client := &dns.Client{}

	// Create query for A record
	msg := &dns.Msg{}
	msg.SetQuestion(dns.Fqdn(hostname), dns.TypeA)

	fmt.Printf("🔍 Querying %s for %s (A record)\n", server, hostname)

	// Send query
	response, _, err := client.Exchange(msg, server)
	if err != nil {
		log.Fatalf("❌ Query failed: %v", err)
	}

	// Display results
	if len(response.Answer) == 0 {
		fmt.Println("❌ No answers received")
		return
	}

	fmt.Printf("✅ Received %d answer(s):\n", len(response.Answer))
	for _, answer := range response.Answer {
		fmt.Printf("   %s\n", answer.String())
	}
}
