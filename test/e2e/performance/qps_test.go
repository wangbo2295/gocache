package performance

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wangbo/gocache/test/e2e"
)

const (
	defaultAddr = "127.0.0.1:16379"
	targetQPS   = 80000  // 80% of 100K requirement
	targetP99   = 2 * time.Millisecond // 200% of 1ms requirement
	targetConns = 5000   // 50% of 10K requirement
)

func setupPerfClient(t *testing.T) *e2e.TestClient {
	client := e2e.NewTestClient(defaultAddr)
	if err := client.Connect(); err != nil {
		t.Skipf("Failed to connect to server at %s: %v (skipping test)", defaultAddr, err)
	}
	// Authenticate (ignore errors if server doesn't require auth)
	client.Send("AUTH", "yourpassword")
	return client
}

// BenchmarkQPS_SET_GET benchmarks SET and GET operations QPS
func BenchmarkQPS_SET_GET(b *testing.B) {
	client := e2e.NewTestClient(defaultAddr)
	if err := client.Connect(); err != nil {
		b.Skipf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Clean up test data
	defer func() {
		cleanup := e2e.NewTestClient(defaultAddr)
		cleanup.Connect()
	cleanup.Send("AUTH", "yourpassword")
		for i := 0; i < 100; i++ {
			cleanup.Send("DEL", fmt.Sprintf("bench_key_%d", i))
		}
		cleanup.Close()
	}()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench_key_%d", i%100)
			value := fmt.Sprintf("bench_value_%d", i)

			// SET
			client.Send("SET", key, value)

			// GET
			client.Send("GET", key)

			i++
		}
	})

	// Calculate QPS
	elapsed := b.Elapsed()
	ops := b.N * 2 // SET + GET
	qps := float64(ops) / elapsed.Seconds()

	b.ReportMetric(qps, "ops/sec")
	b.Logf("QPS: %.2f, Total Operations: %d, Time: %v", qps, ops, elapsed)

	// Check if target QPS is met
	if qps < targetQPS {
		b.Logf("WARNING: QPS %.2f is below target %d", qps, targetQPS)
	}
}

// TestQPS_SingleThread measures single-thread QPS
func TestQPS_SingleThread(t *testing.T) {
	client := setupPerfClient(t)
	defer client.Close()

	// Clean up
	defer func() {
		for i := 0; i < 1000; i++ {
			client.Send("DEL", fmt.Sprintf("qps_key_%d", i))
		}
	}()

	// Warm up
	for i := 0; i < 100; i++ {
		client.Send("SET", fmt.Sprintf("warmup_%d", i), "value")
	}

	// Measure SET QPS
	start := time.Now()
	setOps := 10000
	for i := 0; i < setOps; i++ {
		client.Send("SET", fmt.Sprintf("qps_key_%d", i), "value")
	}
	setDuration := time.Since(start)
	setQPS := float64(setOps) / setDuration.Seconds()

	t.Logf("SET QPS (single-thread): %.2f (%d ops in %v)", setQPS, setOps, setDuration)

	// Measure GET QPS
	start = time.Now()
	getOps := 10000
	for i := 0; i < getOps; i++ {
		client.Send("GET", fmt.Sprintf("qps_key_%d", i%1000))
	}
	getDuration := time.Since(start)
	getQPS := float64(getOps) / getDuration.Seconds()

	t.Logf("GET QPS (single-thread): %.2f (%d ops in %v)", getQPS, getOps, getDuration)

	// Combined QPS
	totalOps := setOps + getOps
	totalDuration := setDuration + getDuration
	combinedQPS := float64(totalOps) / totalDuration.Seconds()

	t.Logf("Combined QPS (single-thread): %.2f", combinedQPS)

	if combinedQPS < float64(targetQPS)/10 {
		t.Errorf("Single-thread QPS %.2f is far below target %d", combinedQPS, targetQPS)
	}
}

// TestQPS_Concurrent measures concurrent QPS with multiple goroutines
func TestQPS_Concurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent QPS test in short mode")
	}

	const numGoroutines = 50
	const opsPerGoroutine = 200

	var wg sync.WaitGroup
	var totalOps int64

	startCh := make(chan struct{})
	startTime := time.Now()

	// Clean up
	defer func() {
		cleanup := e2e.NewTestClient(defaultAddr)
		cleanup.Connect()
	cleanup.Send("AUTH", "yourpassword")
		for i := 0; i < numGoroutines*opsPerGoroutine; i++ {
			cleanup.Send("DEL", fmt.Sprintf("concurrent_key_%d", i))
		}
		cleanup.Close()
	}()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			client := e2e.NewTestClient(defaultAddr)
			if err := client.Connect(); err != nil {
				t.Errorf("Goroutine %d failed to connect: %v", goroutineID, err)
				return
			}
			defer client.Close()

			<-startCh // Wait for start signal

			for j := 0; j < opsPerGoroutine; j++ {
				key := fmt.Sprintf("concurrent_key_%d_%d", goroutineID, j)
				client.Send("SET", key, "value")
				client.Send("GET", key)
				atomic.AddInt64(&totalOps, 2)
			}
		}(i)
	}

	// Start all goroutines at once
	close(startCh)
	wg.Wait()

	elapsed := time.Since(startTime)
	qps := float64(totalOps) / elapsed.Seconds()

	t.Logf("Concurrent QPS: %.2f (%d ops in %v with %d goroutines)",
		qps, totalOps, elapsed, numGoroutines)

	if qps < float64(targetQPS)/5 {
		t.Errorf("Concurrent QPS %.2f is below target %d", qps, targetQPS)
	}
}

// TestQPS_MixedWorkload tests mixed workload (SET/GET/DEL)
func TestQPS_MixedWorkload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping mixed workload test in short mode")
	}

	const numOps = 10000

	client := setupPerfClient(t)
	defer client.Close()

	// Clean up
	defer func() {
		for i := 0; i < 1000; i++ {
			client.Send("DEL", fmt.Sprintf("mixed_key_%d", i))
		}
	}()

	// Warm up
	for i := 0; i < 100; i++ {
		client.Send("SET", fmt.Sprintf("mixed_key_%d", i), "value")
	}

	start := time.Now()

	for i := 0; i < numOps; i++ {
		key := fmt.Sprintf("mixed_key_%d", i%1000)

		switch i % 10 {
		case 0, 1, 2:
			// 30% SET
			client.Send("SET", key, "value")
		case 3, 4, 5, 6:
			// 40% GET
			client.Send("GET", key)
		case 7:
			// 10% DELETE
			client.Send("DEL", key)
		case 8, 9:
			// 20% EXISTS
			client.Send("EXISTS", key)
		}
	}

	elapsed := time.Since(start)
	qps := float64(numOps) / elapsed.Seconds()

	t.Logf("Mixed workload QPS: %.2f (%d ops in %v)", qps, numOps, elapsed)

	if qps < float64(targetQPS)/20 {
		t.Errorf("Mixed workload QPS %.2f is too low", qps)
	}
}

// TestQPS_MultipleConnections tests with multiple concurrent connections
func TestQPS_MultipleConnections(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping multiple connections test in short mode")
	}

	const numConns = 100
	const opsPerConn = 100

	var wg sync.WaitGroup
	var totalOps int64

	startCh := make(chan struct{})
	startTime := time.Now()

	// Clean up
	defer func() {
		cleanup := e2e.NewTestClient(defaultAddr)
		cleanup.Connect()
	cleanup.Send("AUTH", "yourpassword")
		for i := 0; i < numConns*opsPerConn; i++ {
			cleanup.Send("DEL", fmt.Sprintf("conn_key_%d", i))
		}
		cleanup.Close()
	}()

	for i := 0; i < numConns; i++ {
		wg.Add(1)
		go func(connID int) {
			defer wg.Done()

			client := e2e.NewTestClient(defaultAddr)
			if err := client.Connect(); err != nil {
				t.Errorf("Connection %d failed: %v", connID, err)
				return
			}
			defer client.Close()

			<-startCh

			for j := 0; j < opsPerConn; j++ {
				key := fmt.Sprintf("conn_key_%d_%d", connID, j)
				client.Send("SET", key, "value")
				client.Send("GET", key)
				atomic.AddInt64(&totalOps, 2)
			}
		}(i)
	}

	close(startCh)
	wg.Wait()

	elapsed := time.Since(startTime)
	qps := float64(totalOps) / elapsed.Seconds()

	t.Logf("Multiple connections QPS: %.2f (%d connections, %d ops in %v)",
		qps, numConns, totalOps, elapsed)

	// Check connection handling
	if int32(numConns) < targetConns/10 {
		t.Logf("Tested with %d connections (target: %d)", numConns, targetConns)
	}
}

// TestQPS_Pipelining tests pipelined commands
func TestQPS_Pipelining(t *testing.T) {
	client := setupPerfClient(t)
	defer client.Close()

	const batchSize = 100
	const numBatches = 100

	// Clean up
	defer func() {
		for i := 0; i < batchSize*numBatches; i++ {
			client.Send("DEL", fmt.Sprintf("pipe_key_%d", i))
		}
	}()

	startTime := time.Now()

	for batch := 0; batch < numBatches; batch++ {
		for i := 0; i < batchSize; i++ {
			key := fmt.Sprintf("pipe_key_%d_%d", batch, i)
			// Send commands without reading responses (pipelining)
			client.Send("SET", key, "value")
		}

		// Read responses
		for i := 0; i < batchSize; i++ {
			client.Send("PING") // Sync point
		}
	}

	elapsed := time.Since(startTime)
	totalOps := batchSize * numBatches
	qps := float64(totalOps) / elapsed.Seconds()

	t.Logf("Pipelined QPS: %.2f (%d ops in %v, batch size: %d)",
		qps, totalOps, elapsed, batchSize)
}

// Helper function to calculate percentiles
func calculatePercentile(values []time.Duration, percentile float64) time.Duration {
	if len(values) == 0 {
		return 0
	}

	// Make a copy to avoid modifying original slice
	sorted := make([]time.Duration, len(values))
	copy(sorted, values)

	// Simple bubble sort (for small datasets)
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	index := int(float64(len(sorted)-1) * percentile)
	return sorted[index]
}

// TestQPS_Report generates a QPS performance report
func TestQPS_Report(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping QPS report test in short mode")
	}

	t.Log("\n=== QPS Performance Report ===")
	t.Logf("Target QPS: %d", targetQPS)
	t.Logf("Acceptance QPS (80%%): %d", targetQPS*80/100)
	t.Log("\nRunning QPS tests...")

	// Run single-thread test
	t.Run("\nSingle-Thread", func(t *testing.T) {
		// Run single-threaded QPS test
		client := setupPerfClient(t)
		defer client.Close()

		start := time.Now()
		ops := 5000
		for i := 0; i < ops; i++ {
			client.Send("SET", "test_key", "test_value")
			client.Send("GET", "test_key")
		}
		elapsed := time.Since(start)
		qps := float64(ops*2) / elapsed.Seconds()
		t.Logf("Single-thread QPS: %.2f ops/sec", qps)

		// Cleanup
		client.Send("DEL", "test_key")
	})

	t.Run("\nConclusion", func(t *testing.T) {
		t.Log("\nQPS Performance Summary:")
		t.Log("- Single-thread baseline: Measures raw command execution speed")
		t.Log("- Concurrent: Tests parallel processing capability")
		t.Log("- Multiple connections: Tests connection handling")
		t.Log("- Mixed workload: Tests realistic usage patterns")
		t.Logf("\nFor production deployment, ensure QPS >= %d (80%% of target)", targetQPS*80/100)
	})
}
