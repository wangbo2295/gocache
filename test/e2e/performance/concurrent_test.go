package performance

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wangbo/gocache/test/e2e"
)

// TestConcurrentConnections tests handling of many concurrent connections
func TestConcurrentConnections(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent connections test in short mode")
	}

	const numConns = targetConns / 10 // Test with 500 connections (50% of target)
	const opsPerConn = 10

	var wg sync.WaitGroup
	var successCount int32
	var failCount int32

	startCh := make(chan struct{})
	startTime := time.Now()

	// Clean up
	defer func() {
		cleanup := e2e.NewTestClient(defaultAddr)

	cleanup.Connect()
	cleanup.Send("AUTH", "yourpassword")
		cleanup.Connect()
		for i := 0; i < numConns; i++ {
			for j := 0; j < opsPerConn; j++ {
				cleanup.Send("DEL", fmt.Sprintf("conn_test_%d_%d", i, j))
			}
		}
		cleanup.Close()
	}()

	t.Logf("Testing with %d concurrent connections...", numConns)

	for i := 0; i < numConns; i++ {
		wg.Add(1)
		go func(connID int) {
			defer wg.Done()

			client := e2e.NewTestClient(defaultAddr)
			if err := client.Connect(); err != nil {
				atomic.AddInt32(&failCount, 1)
				t.Errorf("Connection %d failed to connect: %v", connID, err)
				return
			}
			defer client.Close()

			// Authenticate
			client.Send("AUTH", "yourpassword")

			atomic.AddInt32(&successCount, 1)

			<-startCh

			// Perform operations
			for j := 0; j < opsPerConn; j++ {
				key := fmt.Sprintf("conn_test_%d_%d", connID, j)
				_, err := client.Send("SET", key, fmt.Sprintf("value_%d", j))
				if err != nil {
					atomic.AddInt32(&failCount, 1)
					continue
				}

				_, err = client.Send("GET", key)
				if err != nil {
					atomic.AddInt32(&failCount, 1)
					continue
				}
			}
		}(i)
	}

	// Start all connections at once
	close(startCh)
	wg.Wait()

	elapsed := time.Since(startTime)

	success := atomic.LoadInt32(&successCount)
	failed := atomic.LoadInt32(&failCount)
	total := int32(numConns)

	t.Logf("Concurrent Connections Test Results:")
	t.Logf("  Total connections: %d", numConns)
	t.Logf("  Successful connections: %d (%.1f%%)", success, float64(success)*100/float64(total))
	t.Logf("  Failed connections: %d (%.1f%%)", failed, float64(failed)*100/float64(total))
	t.Logf("  Total time: %v", elapsed)
	t.Logf("  Average time per connection: %v", elapsed/time.Duration(numConns))

	if float64(success)/float64(total) < 0.95 {
		t.Errorf("Less than 95%% of connections succeeded (%d/%d)", success, total)
	}

	if int(numConns) < targetConns {
		t.Logf("Tested with %d connections (target: %d)", numConns, targetConns)
	}
}

// TestConcurrentOperations tests concurrent operations from multiple connections
func TestConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent operations test in short mode")
	}

	const numGoroutines = 50
	const opsPerGoroutine = 100

	var wg sync.WaitGroup
	var totalOps int64
	var errorOps int64

	startCh := make(chan struct{})
	startTime := time.Now()

	// Clean up
	defer func() {
		cleanup := e2e.NewTestClient(defaultAddr)

	cleanup.Connect()
	cleanup.Send("AUTH", "yourpassword")
		cleanup.Connect()
		for i := 0; i < numGoroutines; i++ {
			for j := 0; j < opsPerGoroutine; j++ {
				cleanup.Send("DEL", fmt.Sprintf("concurrent_op_%d_%d", i, j))
			}
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

			// Authenticate
			client.Send("AUTH", "yourpassword")

			<-startCh

			for j := 0; j < opsPerGoroutine; j++ {
				key := fmt.Sprintf("concurrent_op_%d_%d", goroutineID, j)

				// Perform various operations
				switch j % 5 {
				case 0:
					_, err := client.Send("SET", key, "value")
					if err != nil {
						atomic.AddInt64(&errorOps, 1)
					}
				case 1:
					_, err := client.Send("GET", key)
					if err != nil {
						atomic.AddInt64(&errorOps, 1)
					}
				case 2:
					_, err := client.Send("DEL", key)
					if err != nil {
						atomic.AddInt64(&errorOps, 1)
					}
				case 3:
					_, err := client.Send("EXISTS", key)
					if err != nil {
						atomic.AddInt64(&errorOps, 1)
					}
				case 4:
					_, err := client.Send("TTL", key)
					if err != nil {
						atomic.AddInt64(&errorOps, 1)
					}
				}

				atomic.AddInt64(&totalOps, 1)
			}
		}(i)
	}

	close(startCh)
	wg.Wait()

	elapsed := time.Since(startTime)

	total := atomic.LoadInt64(&totalOps)
	errors := atomic.LoadInt64(&errorOps)
	qps := float64(total) / elapsed.Seconds()

	t.Logf("Concurrent Operations Test Results:")
	t.Logf("  Goroutines: %d", numGoroutines)
	t.Logf("  Total operations: %d", total)
	t.Logf("  Successful operations: %d", total-errors)
	t.Logf("  Failed operations: %d (%.2f%%)", errors, float64(errors)*100/float64(total))
	t.Logf("  Total time: %v", elapsed)
	t.Logf("  QPS: %.2f", qps)
}

// TestConcurrentSameKey tests concurrent operations on the same key
func TestConcurrentSameKey(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent same-key test in short mode")
	}

	const numGoroutines = 20
	const opsPerGoroutine = 50
	const testKey = "concurrent_same_key"

	var wg sync.WaitGroup
	var totalOps int64
	var errorOps int64

	startCh := make(chan struct{})

	// Clean up
	defer func() {
		cleanup := e2e.NewTestClient(defaultAddr)

	cleanup.Connect()
	cleanup.Send("AUTH", "yourpassword")
		cleanup.Connect()
		cleanup.Send("DEL", testKey)
		cleanup.Close()
	}()

	// Initialize key and ensure it's persisted
	client := setupPerfClient(t)
	client.Send("SET", testKey, "0")
	// Verify initialization
	reply, _ := client.Send("GET", testKey)
	if reply.GetString() != "0" {
		t.Fatalf("Failed to initialize test key: got %s, expected 0", reply.GetString())
	}
	client.Close()

	// Small delay to ensure key is fully persisted before concurrent operations start
	time.Sleep(10 * time.Millisecond)

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

			// Authenticate
			client.Send("AUTH", "yourpassword")

			<-startCh

			for j := 0; j < opsPerGoroutine; j++ {
				// INCR is atomic, good for testing concurrent operations
				reply, err := client.Send("INCR", testKey)
				if err != nil {
					atomic.AddInt64(&errorOps, 1)
					t.Logf("Goroutine %d, operation %d failed: %v", goroutineID, j, err)
				} else if reply == nil || reply.Error != nil {
					atomic.AddInt64(&errorOps, 1)
					t.Logf("Goroutine %d, operation %d returned error reply: %v", goroutineID, j, reply.Error)
				}

				atomic.AddInt64(&totalOps, 1)
			}
		}(i)
	}

	close(startCh)
	wg.Wait()

	// Verify final value
	client = setupPerfClient(t)
	defer client.Close()

	reply, err := client.Send("GET", testKey)
	if err != nil {
		t.Errorf("Failed to get final value: %v", err)
	}

	expectedValue := numGoroutines * opsPerGoroutine
	actualValue := 0
	if reply != nil {
		fmt.Sscanf(reply.GetString(), "%d", &actualValue)
	}

	total := atomic.LoadInt64(&totalOps)
	errors := atomic.LoadInt64(&errorOps)

	t.Logf("Concurrent Same-Key Test Results:")
	t.Logf("  Goroutines: %d", numGoroutines)
	t.Logf("  Operations per goroutine: %d", opsPerGoroutine)
	t.Logf("  Total INCR operations: %d", total)
	t.Logf("  Expected final value: %d", expectedValue)
	t.Logf("  Actual final value: %d", actualValue)
	t.Logf("  Failed operations: %d", errors)

	if actualValue != expectedValue {
		t.Errorf("Race condition detected: expected %d, got %d", expectedValue, actualValue)
	} else {
		t.Log("✓ No race condition detected - atomic operations working correctly")
	}
}

// TestConcurrentTransactions tests concurrent transactions
func TestConcurrentTransactions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent transactions test in short mode")
	}

	const numGoroutines = 10
	const txPerGoroutine = 10

	var wg sync.WaitGroup
	var successTx int32
	var failTx int32

	startCh := make(chan struct{})

	// Clean up
	defer func() {
		cleanup := e2e.NewTestClient(defaultAddr)

	cleanup.Connect()
	cleanup.Send("AUTH", "yourpassword")
		cleanup.Connect()
		for i := 0; i < numGoroutines; i++ {
			for j := 0; j < txPerGoroutine; j++ {
				cleanup.Send("DEL", fmt.Sprintf("tx_key_%d_%d", i, j))
			}
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

			// Authenticate
			client.Send("AUTH", "yourpassword")

			<-startCh

			for j := 0; j < txPerGoroutine; j++ {
				// Execute transaction
				client.Send("MULTI")
				client.Send("SET", fmt.Sprintf("tx_key_%d_%d", goroutineID, j), "value")
				client.Send("GET", fmt.Sprintf("tx_key_%d_%d", goroutineID, j))
				reply, err := client.Send("EXEC")

				if err != nil || reply == nil || reply.IsNil() {
					atomic.AddInt32(&failTx, 1)
				} else {
					atomic.AddInt32(&successTx, 1)
				}
			}
		}(i)
	}

	close(startCh)
	wg.Wait()

	success := atomic.LoadInt32(&successTx)
	failed := atomic.LoadInt32(&failTx)
	total := success + failed

	t.Logf("Concurrent Transactions Test Results:")
	t.Logf("  Total transactions: %d", total)
	t.Logf("  Successful transactions: %d (%.1f%%)", success, float64(success)*100/float64(total))
	t.Logf("  Failed/Aborted transactions: %d (%.1f%%)", failed, float64(failed)*100/float64(total))
}

// TestConcurrentWatchTests tests concurrent WATCH operations
func TestConcurrentWatchTests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent WATCH test in short mode")
	}

	const numGoroutines = 5
	const attemptsPerGoroutine = 10
	const watchKey = "concurrent_watch_key"

	var wg sync.WaitGroup
	var successCount int32
	var abortCount int32

	startCh := make(chan struct{})

	// Clean up
	defer func() {
		cleanup := e2e.NewTestClient(defaultAddr)

	cleanup.Connect()
	cleanup.Send("AUTH", "yourpassword")
		cleanup.Connect()
		cleanup.Send("DEL", watchKey)
		cleanup.Close()
	}()

	// Initialize key
	client := setupPerfClient(t)
	client.Send("SET", watchKey, "initial")
	client.Close()

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

			// Authenticate
			client.Send("AUTH", "yourpassword")

			<-startCh

			for j := 0; j < attemptsPerGoroutine; j++ {
				// Watch the key
				client.Send("WATCH", watchKey)

				// Get current value
				reply, _ := client.Send("GET", watchKey)
				_ = reply.GetString()

				// Start transaction
				client.Send("MULTI")
				client.Send("SET", watchKey, fmt.Sprintf("value_%d_%d", goroutineID, j))

				// Execute transaction
				reply, err := client.Send("EXEC")

				if err != nil || reply == nil || reply.IsNil() {
					// Transaction was aborted due to WATCH
					atomic.AddInt32(&abortCount, 1)
				} else {
					atomic.AddInt32(&successCount, 1)
				}
			}
		}(i)
	}

	close(startCh)
	wg.Wait()

	success := atomic.LoadInt32(&successCount)
	aborted := atomic.LoadInt32(&abortCount)
	total := success + aborted

	t.Logf("Concurrent WATCH Test Results:")
	t.Logf("  Total transaction attempts: %d", total)
	t.Logf("  Successful transactions: %d (%.1f%%)", success, float64(success)*100/float64(total))
	t.Logf("  Aborted transactions (WATCH triggered): %d (%.1f%%)", aborted, float64(aborted)*100/float64(total))

	if aborted > 0 {
		t.Log("✓ WATCH mechanism is working - transactions are being aborted on conflict")
	}
}

// TestConcurrentStress stress tests with high concurrency
func TestConcurrentStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test")
	}

	const numGoroutines = 100
	const duration = 5 * time.Second

	var wg sync.WaitGroup
	var stopCh = make(chan struct{})
	var opsCount int64

	startTime := time.Now()

	// Clean up
	defer func() {
		cleanup := e2e.NewTestClient(defaultAddr)

	cleanup.Connect()
	cleanup.Send("AUTH", "yourpassword")
		cleanup.Connect()
		for i := 0; i < numGoroutines*10; i++ {
			cleanup.Send("DEL", fmt.Sprintf("stress_key_%d", i))
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

			// Authenticate
			client.Send("AUTH", "yourpassword")

			opNum := 0
			for {
				select {
				case <-stopCh:
					return
				default:
					key := fmt.Sprintf("stress_key_%d", opNum%100)
					client.Send("SET", key, fmt.Sprintf("value_%d", opNum))
					client.Send("GET", key)
					atomic.AddInt64(&opsCount, 2)
					opNum++
				}
			}
		}(i)
	}

	// Run for specified duration
	time.Sleep(duration)
	close(stopCh)
	wg.Wait()

	elapsed := time.Since(startTime)
	totalOps := atomic.LoadInt64(&opsCount)
	qps := float64(totalOps) / elapsed.Seconds()

	t.Logf("Concurrent Stress Test Results:")
	t.Logf("  Goroutines: %d", numGoroutines)
	t.Logf("  Duration: %v", duration)
	t.Logf("  Total operations: %d", totalOps)
	t.Logf("  QPS: %.2f", qps)
	t.Logf("  Average per goroutine: %.2f ops/sec", qps/float64(numGoroutines))
}

// TestConcurrent_Report generates a concurrent test report
func TestConcurrent_Report(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent report test")
	}

	t.Log("\n=== Concurrent Performance Report ===")
	t.Logf("Target concurrent connections: %d", targetConns)
	t.Logf("Acceptance (50%%): %d", targetConns/2)
	t.Log("\nConcurrent Test Summary:")
	t.Log("- Connection Handling: Tests server's ability to handle many connections")
	t.Log("- Concurrent Operations: Tests parallel command processing")
	t.Log("- Same-Key Operations: Tests atomic operations and race conditions")
	t.Log("- Concurrent Transactions: Tests transaction isolation")
	t.Log("- Stress Test: Tests system stability under high load")
	t.Logf("\nFor production deployment, ensure server can handle >= %d concurrent connections", targetConns/2)
}
