package performance

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/wangbo/gocache/test/e2e"
)

// TestLatency_BasicOperations measures latency of basic operations
func TestLatency_BasicOperations(t *testing.T) {
	client := setupPerfClient(t)
	defer client.Close()

	// Clean up
	defer client.Send("DEL", "latency_test_key")

	t.Run("SET latency", func(t *testing.T) {
		latencies := make([]time.Duration, 0, 1000)

		// Warm up
		for i := 0; i < 100; i++ {
			client.Send("SET", "warmup", "value")
		}

		// Measure SET latency
		for i := 0; i < 1000; i++ {
			start := time.Now()
			client.Send("SET", "latency_test_key", "value")
			elapsed := time.Since(start)
			latencies = append(latencies, elapsed)
		}

		// Calculate percentiles
		sort.Slice(latencies, func(i, j int) bool {
			return latencies[i] < latencies[j]
		})

		p50 := latencies[len(latencies)*50/100]
		p95 := latencies[len(latencies)*95/100]
		p99 := latencies[len(latencies)*99/100]
		p999 := latencies[len(latencies)*999/1000]
		avg := average(latencies)

		t.Logf("SET Latency (1000 ops):")
		t.Logf("  Average: %v", avg)
		t.Logf("  P50: %v", p50)
		t.Logf("  P95: %v", p95)
		t.Logf("  P99: %v", p99)
		t.Logf("  P99.9: %v", p999)

		// Check if P99 meets target
		if p99 > targetP99 {
			t.Logf("WARNING: P99 latency %v exceeds target %v", p99, targetP99)
		}
	})

	t.Run("GET latency", func(t *testing.T) {
		client.Send("SET", "latency_test_key", "value")

		latencies := make([]time.Duration, 0, 1000)

		// Measure GET latency
		for i := 0; i < 1000; i++ {
			start := time.Now()
			client.Send("GET", "latency_test_key")
			elapsed := time.Since(start)
			latencies = append(latencies, elapsed)
		}

		sort.Slice(latencies, func(i, j int) bool {
			return latencies[i] < latencies[j]
		})

		p50 := latencies[len(latencies)*50/100]
		p95 := latencies[len(latencies)*95/100]
		p99 := latencies[len(latencies)*99/100]
		p999 := latencies[len(latencies)*999/1000]
		avg := average(latencies)

		t.Logf("GET Latency (1000 ops):")
		t.Logf("  Average: %v", avg)
		t.Logf("  P50: %v", p50)
		t.Logf("  P95: %v", p95)
		t.Logf("  P99: %v", p99)
		t.Logf("  P99.9: %v", p999)

		if p99 > targetP99 {
			t.Logf("WARNING: P99 latency %v exceeds target %v", p99, targetP99)
		}
	})

	// Cleanup
	client.Send("DEL", "latency_test_key")
}

// TestLatency_AllDataTypes measures latency across different data types
func TestLatency_AllDataTypes(t *testing.T) {
	client := setupPerfClient(t)
	defer client.Close()

	// Clean up
	defer func() {
		client.Send("DEL", "latency_string")
		client.Send("DEL", "latency_hash")
		client.Send("DEL", "latency_list")
		client.Send("DEL", "latency_set")
		client.Send("DEL", "latency_zset")
	}()

	const iterations = 500

	t.Run("String operations latency", func(t *testing.T) {
		latencies := make([]time.Duration, 0, iterations)

		for i := 0; i < iterations; i++ {
			start := time.Now()
			client.Send("SET", "latency_string", "value")
			elapsed := time.Since(start)
			latencies = append(latencies, elapsed)
		}

		sort.Slice(latencies, func(i, j int) bool {
			return latencies[i] < latencies[j]
		})

		p99 := latencies[len(latencies)*99/100]
		avg := average(latencies)

		t.Logf("String SET - Avg: %v, P99: %v", avg, p99)
	})

	t.Run("Hash operations latency", func(t *testing.T) {
		latencies := make([]time.Duration, 0, iterations)

		for i := 0; i < iterations; i++ {
			start := time.Now()
			client.Send("HSET", "latency_hash", "field", "value")
			elapsed := time.Since(start)
			latencies = append(latencies, elapsed)
		}

		sort.Slice(latencies, func(i, j int) bool {
			return latencies[i] < latencies[j]
		})

		p99 := latencies[len(latencies)*99/100]
		avg := average(latencies)

		t.Logf("Hash HSET - Avg: %v, P99: %v", avg, p99)
	})

	t.Run("List operations latency", func(t *testing.T) {
		latencies := make([]time.Duration, 0, iterations)

		for i := 0; i < iterations; i++ {
			start := time.Now()
			client.Send("LPUSH", "latency_list", "value")
			elapsed := time.Since(start)
			latencies = append(latencies, elapsed)
		}

		sort.Slice(latencies, func(i, j int) bool {
			return latencies[i] < latencies[j]
		})

		p99 := latencies[len(latencies)*99/100]
		avg := average(latencies)

		t.Logf("List LPUSH - Avg: %v, P99: %v", avg, p99)
	})

	t.Run("Set operations latency", func(t *testing.T) {
		latencies := make([]time.Duration, 0, iterations)

		for i := 0; i < iterations; i++ {
			start := time.Now()
			client.Send("SADD", "latency_set", "member")
			elapsed := time.Since(start)
			latencies = append(latencies, elapsed)
		}

		sort.Slice(latencies, func(i, j int) bool {
			return latencies[i] < latencies[j]
		})

		p99 := latencies[len(latencies)*99/100]
		avg := average(latencies)

		t.Logf("Set SADD - Avg: %v, P99: %v", avg, p99)
	})

	t.Run("SortedSet operations latency", func(t *testing.T) {
		latencies := make([]time.Duration, 0, iterations)

		for i := 0; i < iterations; i++ {
			start := time.Now()
			client.Send("ZADD", "latency_zset", "1", "member")
			elapsed := time.Since(start)
			latencies = append(latencies, elapsed)
		}

		sort.Slice(latencies, func(i, j int) bool {
			return latencies[i] < latencies[j]
		})

		p99 := latencies[len(latencies)*99/100]
		avg := average(latencies)

		t.Logf("SortedSet ZADD - Avg: %v, P99: %v", avg, p99)
	})
}

// TestLatency_Concurrent measures latency under concurrent load
func TestLatency_Concurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent latency test in short mode")
	}

	const numGoroutines = 10
	const opsPerGoroutine = 100

	client := setupPerfClient(t)
	defer client.Close()

	// Clean up
	defer func() {
		for i := 0; i < numGoroutines*opsPerGoroutine; i++ {
			client.Send("DEL", fmt.Sprintf("concurrent_latency_%d", i))
		}
	}()

	allLatencies := make(chan time.Duration, numGoroutines*opsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			client := e2e.NewTestClient(defaultAddr)
			client.Connect()
			defer client.Close()

			for j := 0; j < opsPerGoroutine; j++ {
				start := time.Now()
				client.Send("SET", fmt.Sprintf("concurrent_latency_%d_%d", goroutineID, j), "value")
				elapsed := time.Since(start)
				allLatencies <- elapsed
			}
		}(i)
	}

	// Collect all latencies
	latencies := make([]time.Duration, 0, numGoroutines*opsPerGoroutine)
	for i := 0; i < numGoroutines*opsPerGoroutine; i++ {
		latency := <-allLatencies
		latencies = append(latencies, latency)
	}

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	p50 := latencies[len(latencies)*50/100]
	p95 := latencies[len(latencies)*95/100]
	p99 := latencies[len(latencies)*99/100]
	p999 := latencies[len(latencies)*999/1000]
	avg := average(latencies)

	t.Logf("Concurrent Latency (%d goroutines, %d total ops):", numGoroutines, len(latencies))
	t.Logf("  Average: %v", avg)
	t.Logf("  P50: %v", p50)
	t.Logf("  P95: %v", p95)
	t.Logf("  P99: %v", p99)
	t.Logf("  P99.9: %v", p999)

	if p99 > targetP99*2 {
		t.Logf("WARNING: Concurrent P99 latency %v exceeds 2x target %v", p99, targetP99)
	}
}

// TestLatency_WithTTL measures latency with TTL operations
func TestLatency_WithTTL(t *testing.T) {
	client := setupPerfClient(t)
	defer client.Close()

	// Clean up
	defer client.Send("DEL", "ttl_latency_key")

	const iterations = 500

	t.Run("EXPIRE latency", func(t *testing.T) {
		client.Send("SET", "ttl_latency_key", "value")

		latencies := make([]time.Duration, 0, iterations)

		for i := 0; i < iterations; i++ {
			start := time.Now()
			client.Send("EXPIRE", "ttl_latency_key", "3600")
			elapsed := time.Since(start)
			latencies = append(latencies, elapsed)
		}

		sort.Slice(latencies, func(i, j int) bool {
			return latencies[i] < latencies[j]
		})

		p99 := latencies[len(latencies)*99/100]
		avg := average(latencies)

		t.Logf("EXPIRE - Avg: %v, P99: %v", avg, p99)
	})

	t.Run("TTL latency", func(t *testing.T) {
		latencies := make([]time.Duration, 0, iterations)

		for i := 0; i < iterations; i++ {
			start := time.Now()
			client.Send("TTL", "ttl_latency_key")
			elapsed := time.Since(start)
			latencies = append(latencies, elapsed)
		}

		sort.Slice(latencies, func(i, j int) bool {
			return latencies[i] < latencies[j]
		})

		p99 := latencies[len(latencies)*99/100]
		avg := average(latencies)

		t.Logf("TTL - Avg: %v, P99: %v", avg, p99)
	})
}

// TestLatency_Transaction measures transaction latency
func TestLatency_Transaction(t *testing.T) {
	client := setupPerfClient(t)
	defer client.Close()

	// Clean up
	defer func() {
		client.Send("DEL", "tx_latency_key1", "tx_latency_key2", "tx_latency_key3")
	}()

	const iterations = 200

	latencies := make([]time.Duration, 0, iterations)

	for i := 0; i < iterations; i++ {
		start := time.Now()

		client.Send("MULTI")
		client.Send("SET", "tx_latency_key1", "value1")
		client.Send("SET", "tx_latency_key2", "value2")
		client.Send("SET", "tx_latency_key3", "value3")
		client.Send("EXEC")

		elapsed := time.Since(start)
		latencies = append(latencies, elapsed)
	}

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	p50 := latencies[len(latencies)*50/100]
	p95 := latencies[len(latencies)*95/100]
	p99 := latencies[len(latencies)*99/100]
	avg := average(latencies)

	t.Logf("Transaction (3 commands) Latency:")
	t.Logf("  Average: %v", avg)
	t.Logf("  P50: %v", p50)
	t.Logf("  P95: %v", p95)
	t.Logf("  P99: %v", p99)

	// Per-command latency
	avgPerCmd := avg / 3
	t.Logf("  Average per command: %v", avgPerCmd)
}

// TestLatency_Report generates a comprehensive latency report
func TestLatency_Report(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping latency report test in short mode")
	}

	t.Log("\n=== Latency Performance Report ===")
	t.Logf("Target P99: %v", targetP99)
	t.Logf("Acceptance P99 (200%%): %v", targetP99*2)
	t.Log("\nRunning latency tests...")

	client := setupPerfClient(t)
	defer client.Close()

	// Quick latency test
	const quickIterations = 100
	latencies := make([]time.Duration, 0, quickIterations)

	client.Send("SET", "report_key", "value")

	for i := 0; i < quickIterations; i++ {
		start := time.Now()
		client.Send("GET", "report_key")
		elapsed := time.Since(start)
		latencies = append(latencies, elapsed)
	}

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	p50 := latencies[len(latencies)*50/100]
	p95 := latencies[len(latencies)*95/100]
	p99 := latencies[len(latencies)*99/100]
	avg := average(latencies)

	t.Logf("\nQuick Latency Test (GET operation):")
	t.Logf("  Average: %v", avg)
	t.Logf("  P50: %v", p50)
	t.Logf("  P95: %v", p95)
	t.Logf("  P99: %v", p99)

	if p99 > targetP99 {
		t.Logf("WARNING: P99 latency %v exceeds target %v", p99, targetP99)
	} else {
		t.Logf("✓ P99 latency meets target")
	}

	// Cleanup
	client.Send("DEL", "report_key")

	t.Run("\nConclusion", func(t *testing.T) {
		t.Log("\nLatency Performance Summary:")
		t.Log("- P50 (Median): Typical latency experienced by 50% of requests")
		t.Log("- P95: 95% of requests complete within this time")
		t.Log("- P99: 99% of requests complete within this time (SLA target)")
		t.Log("- P99.9: Worst-case latency, 99.9% of requests")
		t.Logf("\nFor production deployment, ensure P99 <= %v", targetP99)
	})
}

// Helper function to calculate average
func average(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	var sum time.Duration
	for _, lat := range latencies {
		sum += lat
	}

	return sum / time.Duration(len(latencies))
}
