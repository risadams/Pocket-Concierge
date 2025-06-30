package main

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/miekg/dns"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run cmd/loadtest/main.go <server:port> <duration-seconds> <concurrent>")
		fmt.Println("Example: go run cmd/loadtest/main.go 127.0.0.1:8053 30 50")
		os.Exit(1)
	}

	server := os.Args[1]
	duration, _ := strconv.Atoi(os.Args[2])
	concurrent, _ := strconv.Atoi(os.Args[3])

	fmt.Printf("🔥 DNS Load Test\n")
	fmt.Printf("📊 Server: %s\n", server)
	fmt.Printf("📊 Duration: %d seconds\n", duration)
	fmt.Printf("📊 Concurrent Workers: %d\n", concurrent)
	fmt.Println("==========================================")

	runLoadTest(server, time.Duration(duration)*time.Second, concurrent)
}

func runLoadTest(server string, duration time.Duration, concurrent int) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	totalQueries := 0
	successfulQueries := 0
	totalLatency := time.Duration(0)

	stopChan := make(chan bool)

	// Start timer
	go func() {
		time.Sleep(duration)
		close(stopChan)
	}()

	// Progress reporter
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-stopChan:
				return
			case <-ticker.C:
				mu.Lock()
				qps := float64(totalQueries) / time.Since(time.Now().Add(-duration)).Seconds()
				successRate := float64(successfulQueries) / float64(totalQueries) * 100
				avgLatency := totalLatency / time.Duration(totalQueries)
				fmt.Printf("\r🔄 Queries: %d | QPS: %.1f | Success: %.1f%% | Avg Latency: %v",
					totalQueries, qps, successRate, avgLatency)
				mu.Unlock()
			}
		}
	}()

	// Launch workers
	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			client := &dns.Client{Timeout: 5 * time.Second}

			queries := []string{
				"ris-desktop.home", "google.com",
				"ris-laptop.home", "github.com",
				"homeserver.home", "stackoverflow.com",
			}

			queryIndex := 0

			for {
				select {
				case <-stopChan:
					return
				default:
					query := queries[queryIndex%len(queries)]
					queryIndex++

					msg := &dns.Msg{}
					msg.SetQuestion(dns.Fqdn(query), dns.TypeA)

					start := time.Now()
					_, _, err := client.Exchange(msg, server)
					latency := time.Since(start)

					mu.Lock()
					totalQueries++
					totalLatency += latency
					if err == nil {
						successfulQueries++
					}
					mu.Unlock()
				}
			}
		}(i)
	}

	wg.Wait()

	fmt.Printf("\n✅ Load test completed!\n")
	fmt.Printf("📊 Total Queries: %d\n", totalQueries)
	fmt.Printf("📊 Successful: %d (%.1f%%)\n", successfulQueries,
		float64(successfulQueries)/float64(totalQueries)*100)
	fmt.Printf("📊 Average QPS: %.2f\n", float64(totalQueries)/duration.Seconds())
	fmt.Printf("📊 Average Latency: %v\n", totalLatency/time.Duration(totalQueries))
}
