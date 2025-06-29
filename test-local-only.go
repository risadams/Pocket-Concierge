package main

import (
	"fmt"
	"log"
	"time"

	"github.com/miekg/dns"
)

func main() {
	client := &dns.Client{
		Timeout: 1 * time.Second,
	}

	localHosts := []string{
		"ris-desktop.home",
		"ris-laptop.home",
		"homeserver.home",
		"nas.home",
		"ipv6-only.home",
	}

	fmt.Println("🔍 Testing local resolution only...")

	for i := 0; i < 10; i++ {
		for _, host := range localHosts {
			start := time.Now()

			msg := &dns.Msg{}
			msg.SetQuestion(dns.Fqdn(host), dns.TypeA)

			_, _, err := client.Exchange(msg, "127.0.0.1:8053")
			latency := time.Since(start)

			if err != nil {
				log.Printf("❌ %s: %v", host, err)
			} else {
				fmt.Printf("✅ %s: %v\n", host, latency)
			}

			if latency > 10*time.Millisecond {
				fmt.Printf("🐌 SLOW: %s took %v\n", host, latency)
			}
		}
		fmt.Printf("--- Round %d complete ---\n", i+1)
		time.Sleep(100 * time.Millisecond)
	}
}
