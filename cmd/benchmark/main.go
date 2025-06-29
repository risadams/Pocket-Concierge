package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/miekg/dns"
)

type BenchmarkResult struct {
	Query        string
	Server       string
	Protocol     string
	Latency      time.Duration
	Success      bool
	ResponseSize int
	Error        string
}

type BenchmarkStats struct {
	TotalQueries      int
	SuccessfulQueries int
	FailedQueries     int
	TotalTime         time.Duration
	MinLatency        time.Duration
	MaxLatency        time.Duration
	AvgLatency        time.Duration
	MedianLatency     time.Duration
	P95Latency        time.Duration
	P99Latency        time.Duration
	QPS               float64
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run cmd/benchmark/main.go <server:port> <queries> <concurrent> [test-type]")
		fmt.Println("Examples:")
		fmt.Println("  go run cmd/benchmark/main.go 127.0.0.1:8053 1000 10 local")
		fmt.Println("  go run cmd/benchmark/main.go 127.0.0.1:8053 1000 10 upstream")
		fmt.Println("  go run cmd/benchmark/main.go 127.0.0.1:8053 1000 10 mixed")
		fmt.Println("  go run cmd/benchmark/main.go 8.8.8.8:53 1000 10 baseline")
		os.Exit(1)
	}

	server := os.Args[1]
	queries, _ := strconv.Atoi(os.Args[2])
	concurrent, _ := strconv.Atoi(os.Args[3])
	testType := "mixed"
	if len(os.Args) > 4 {
		testType = os.Args[4]
	}

	fmt.Printf("🚀 DNS Benchmark Tool\n")
	fmt.Printf("📊 Server: %s\n", server)
	fmt.Printf("📊 Total Queries: %d\n", queries)
	fmt.Printf("📊 Concurrent Workers: %d\n", concurrent)
	fmt.Printf("📊 Test Type: %s\n", testType)
	fmt.Println("==========================================")

	// Run benchmark
	results := runBenchmark(server, queries, concurrent, testType)

	// Calculate and display stats
	stats := calculateStats(results)
	displayResults(stats, server, testType)
}

func runBenchmark(server string, totalQueries, concurrent int, testType string) []BenchmarkResult {
	var results []BenchmarkResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Create work queue
	workQueue := make(chan string, totalQueries)

	// Fill work queue with test queries
	testQueries := getTestQueries(testType)
	for i := 0; i < totalQueries; i++ {
		workQueue <- testQueries[i%len(testQueries)]
	}
	close(workQueue)

	// Progress tracking
	completed := 0
	progressTicker := time.NewTicker(time.Second)
	defer progressTicker.Stop()

	go func() {
		for range progressTicker.C {
			mu.Lock()
			current := completed
			mu.Unlock()

			if current >= totalQueries {
				return
			}

			progress := float64(current) / float64(totalQueries) * 100
			fmt.Printf("\r🔄 Progress: %d/%d (%.1f%%) ", current, totalQueries, progress)
		}
	}()

	// Start benchmark
	startTime := time.Now()

	// Launch workers
	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &dns.Client{Timeout: 10 * time.Second}

			for query := range workQueue {
				result := performQuery(client, server, query)

				mu.Lock()
				results = append(results, result)
				completed++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	totalTime := time.Since(startTime)

	fmt.Printf("\r✅ Completed %d queries in %v\n", totalQueries, totalTime)
	return results
}

func getTestQueries(testType string) []string {
	switch testType {
	case "local":
		return []string{
			"ris-desktop.home",
			"ris-laptop.home",
			"homeserver.home",
			"nas.home",
			"ipv6-only.home",
			"router.home",
			"printer.home",
			"camera.home",
			"switch.home",
		}
	case "upstream":
		return []string{
			"google.com",
			"github.com",
			"stackoverflow.com",
			"reddit.com",
			"youtube.com",
			"cloudflare.com",
			"amazon.com",
			"microsoft.com",
			"facebook.com",
			"twitter.com",
		}
	case "mixed":
		return []string{
			"ris-desktop.home",
			"google.com",
			"ris-laptop.home",
			"github.com",
			"homeserver.home",
			"stackoverflow.com",
			"nas.home",
			"reddit.com",
			"router.home",
			"youtube.com",
		}
	case "baseline":
		return []string{
			"google.com",
			"github.com",
			"stackoverflow.com",
			"reddit.com",
			"youtube.com",
		}
	default:
		return []string{"google.com"}
	}
}

func performQuery(client *dns.Client, server, hostname string) BenchmarkResult {
	msg := &dns.Msg{}
	msg.SetQuestion(dns.Fqdn(hostname), dns.TypeA)

	start := time.Now()
	response, _, err := client.Exchange(msg, server)
	latency := time.Since(start)

	result := BenchmarkResult{
		Query:    hostname,
		Server:   server,
		Protocol: "udp",
		Latency:  latency,
		Success:  err == nil && response != nil,
	}

	if err != nil {
		result.Error = err.Error()
	} else if response != nil {
		result.ResponseSize = response.Len()
	}

	return result
}

func calculateStats(results []BenchmarkResult) BenchmarkStats {
	if len(results) == 0 {
		return BenchmarkStats{}
	}

	var latencies []time.Duration
	var totalLatency time.Duration
	successCount := 0

	for _, result := range results {
		if result.Success {
			successCount++
			latencies = append(latencies, result.Latency)
			totalLatency += result.Latency
		}
	}

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	stats := BenchmarkStats{
		TotalQueries:      len(results),
		SuccessfulQueries: successCount,
		FailedQueries:     len(results) - successCount,
	}

	if len(latencies) > 0 {
		stats.MinLatency = latencies[0]
		stats.MaxLatency = latencies[len(latencies)-1]
		stats.AvgLatency = totalLatency / time.Duration(len(latencies))
		stats.MedianLatency = latencies[len(latencies)/2]
		stats.P95Latency = latencies[int(float64(len(latencies))*0.95)]
		stats.P99Latency = latencies[int(float64(len(latencies))*0.99)]

		// Calculate QPS based on successful queries
		if totalLatency > 0 {
			stats.QPS = float64(successCount) / totalLatency.Seconds()
		}
	}

	return stats
}

func displayResults(stats BenchmarkStats, server, testType string) {
	fmt.Println("\n📊 BENCHMARK RESULTS")
	fmt.Println("==========================================")
	fmt.Printf("🎯 Server: %s\n", server)
	fmt.Printf("🎯 Test Type: %s\n", testType)
	fmt.Printf("📈 Total Queries: %d\n", stats.TotalQueries)
	fmt.Printf("✅ Successful: %d (%.1f%%)\n", stats.SuccessfulQueries,
		float64(stats.SuccessfulQueries)/float64(stats.TotalQueries)*100)
	fmt.Printf("❌ Failed: %d (%.1f%%)\n", stats.FailedQueries,
		float64(stats.FailedQueries)/float64(stats.TotalQueries)*100)

	if stats.SuccessfulQueries > 0 {
		fmt.Println("\n⏱️  LATENCY STATISTICS")
		fmt.Printf("├─ Min:     %v\n", stats.MinLatency)
		fmt.Printf("├─ Max:     %v\n", stats.MaxLatency)
		fmt.Printf("├─ Average: %v\n", stats.AvgLatency)
		fmt.Printf("├─ Median:  %v\n", stats.MedianLatency)
		fmt.Printf("├─ 95th%%:   %v\n", stats.P95Latency)
		fmt.Printf("└─ 99th%%:   %v\n", stats.P99Latency)

		fmt.Println("\n🚀 THROUGHPUT")
		fmt.Printf("└─ QPS: %.2f queries/second\n", stats.QPS)
	}

	fmt.Println("==========================================")
}
