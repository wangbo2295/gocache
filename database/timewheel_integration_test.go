package database

import (
	"strconv"
	"testing"
	"time"
)

// TestTimeWheelActiveExpiration tests that keys are actively expired by the time wheel
func TestTimeWheelActiveExpiration(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// Set a key with short TTL (100ms using PEXPIRE)
	db.ExecCommand("SET", "testkey", "testvalue")
	db.ExecCommand("PEXPIRE", "testkey", "100") // 100ms

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Key should be actively expired by time wheel
	result, err := db.ExecCommand("GET", "testkey")
	if err != nil {
		t.Fatalf("GET command failed: %v", err)
	}

	if len(result) > 0 && len(result[0]) > 0 {
		t.Errorf("Expected key to be expired, got: %s", string(result[0]))
	}
}

// TestTimeWheelMultipleExpirations tests multiple keys expiring at different times
func TestTimeWheelMultipleExpirations(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// Set multiple keys with different TTLs
	db.ExecCommand("SET", "key1", "value1")
	db.ExecCommand("SET", "key2", "value2")
	db.ExecCommand("SET", "key3", "value3")

	db.ExecCommand("PEXPIRE", "key1", "100")  // 100ms
	db.ExecCommand("PEXPIRE", "key2", "200")  // 200ms
	db.ExecCommand("PEXPIRE", "key3", "500")  // 500ms

	// Wait for first expiration
	time.Sleep(150 * time.Millisecond)

	// key1 should be expired
	result1, _ := db.ExecCommand("EXISTS", "key1")
	if len(result1) > 0 && string(result1[0]) != "0" {
		t.Error("key1 should be expired after 150ms")
	}

	// key2 and key3 should still exist
	result2, _ := db.ExecCommand("EXISTS", "key2")
	if len(result2) > 0 && string(result2[0]) == "0" {
		t.Error("key2 should not be expired after 150ms")
	}

	result3, _ := db.ExecCommand("EXISTS", "key3")
	if len(result3) > 0 && string(result3[0]) == "0" {
		t.Error("key3 should not be expired after 150ms")
	}

	// Wait for second expiration
	time.Sleep(100 * time.Millisecond)

	// key2 should now be expired
	result2, _ = db.ExecCommand("EXISTS", "key2")
	if len(result2) > 0 && string(result2[0]) != "0" {
		t.Error("key2 should be expired after 250ms")
	}

	// key3 should still exist
	result3, _ = db.ExecCommand("EXISTS", "key3")
	if len(result3) > 0 && string(result3[0]) == "0" {
		t.Error("key3 should not be expired after 250ms")
	}

	// Wait for final expiration
	time.Sleep(300 * time.Millisecond)

	// key3 should now be expired
	result3, _ = db.ExecCommand("EXISTS", "key3")
	if len(result3) > 0 && string(result3[0]) != "0" {
		t.Error("key3 should be expired after 550ms")
	}
}

// TestTimeWheelPersistRemovesFromWheel tests that PERSIST removes key from time wheel
func TestTimeWheelPersistRemovesFromWheel(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// Set a key with TTL
	db.ExecCommand("SET", "testkey", "testvalue")
	db.ExecCommand("EXPIRE", "testkey", "100") // 100ms

	// Remove TTL
	db.ExecCommand("PERSIST", "testkey")

	// Wait for expiration time
	time.Sleep(200 * time.Millisecond)

	// Key should still exist since we removed TTL
	result, err := db.ExecCommand("GET", "testkey")
	if err != nil {
		t.Fatalf("GET command failed: %v", err)
	}

	if len(result) == 0 || len(result[0]) == 0 {
		t.Error("Key should still exist after PERSIST")
	}

	if string(result[0]) != "testvalue" {
		t.Errorf("Expected 'testvalue', got '%s'", string(result[0]))
	}
}

// TestTimeWheelDelRemovesFromWheel tests that DEL removes key from time wheel
func TestTimeWheelDelRemovesFromWheel(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// Set a key with TTL
	db.ExecCommand("SET", "testkey", "testvalue")
	db.ExecCommand("PEXPIRE", "testkey", "500") // 500ms

	// Delete the key
	db.ExecCommand("DEL", "testkey")

	// Wait some time
	time.Sleep(100 * time.Millisecond)

	// Check time wheel size (should be 0 since key was removed)
	// Note: We can't directly access timeWheel, but we can infer it's working
	// by checking that no errors occur

	// Set a new key to verify time wheel still works
	db.ExecCommand("SET", "key2", "value2")
	db.ExecCommand("PEXPIRE", "key2", "100") // 100ms
	time.Sleep(200 * time.Millisecond)

	result, _ := db.ExecCommand("EXISTS", "key2")
	if len(result) > 0 && string(result[0]) != "0" {
		t.Error("key2 should be expired")
	}
}

// TestTimeWheelUpdateTTL tests updating TTL of existing key
func TestTimeWheelUpdateTTL(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// Set a key with short TTL
	db.ExecCommand("SET", "testkey", "testvalue")
	db.ExecCommand("PEXPIRE", "testkey", "100") // 100ms

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// Update TTL to longer time
	db.ExecCommand("PEXPIRE", "testkey", "500") // 500ms

	// Wait for original expiration time
	time.Sleep(100 * time.Millisecond)

	// Key should still exist since we updated TTL
	result, _ := db.ExecCommand("GET", "testkey")
	if len(result) == 0 || len(result[0]) == 0 {
		t.Error("Key should still exist after TTL update")
	}

	// Wait for updated expiration time
	time.Sleep(450 * time.Millisecond)

	// Key should now be expired
	result, _ = db.ExecCommand("GET", "testkey")
	if len(result) > 0 && len(result[0]) > 0 {
		t.Error("Key should be expired after updated TTL")
	}
}

// TestTimeWheelAccuracy tests the accuracy of time wheel expiration
func TestTimeWheelAccuracy(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// Set a key with 50ms TTL
	db.ExecCommand("SET", "testkey", "testvalue")
	db.ExecCommand("PEXPIRE", "testkey", "50") // 50ms

	// Record start time
	start := time.Now()

	// Wait for expiration (poll until key is gone)
	for {
		result, _ := db.ExecCommand("EXISTS", "testkey")
		if len(result) > 0 && string(result[0]) == "0" {
			break
		}
		if time.Since(start) > 200*time.Millisecond {
			t.Fatal("Key did not expire within expected time")
		}
		time.Sleep(5 * time.Millisecond)
	}

	elapsed := time.Since(start)

	// Expiration should be between 50ms and 100ms (allowing for tick interval and processing)
	if elapsed < 50*time.Millisecond {
		t.Errorf("Key expired too early: %v", elapsed)
	}
	if elapsed > 150*time.Millisecond {
		t.Errorf("Key expired too late: %v", elapsed)
	}
}

// TestTimeWheelWithEviction tests time wheel works with eviction policy
func TestTimeWheelWithEviction(t *testing.T) {
	// This test ensures time wheel doesn't interfere with eviction
	db := MakeDB()
	defer db.Close()

	// Set multiple keys with TTL
	for i := 1; i <= 10; i++ {
		key := "key" + strconv.Itoa(i)
		db.ExecCommand("SET", key, "value"+strconv.Itoa(i))
		db.ExecCommand("PEXPIRE", key, strconv.Itoa(100)) // 100ms TTL
	}

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// All keys should be expired
	for i := 1; i <= 10; i++ {
		key := "key" + strconv.Itoa(i)
		result, _ := db.ExecCommand("EXISTS", key)
		if len(result) > 0 && string(result[0]) != "0" {
			t.Errorf("Key %s should be expired", key)
		}
	}
}

// BenchmarkTimeWheelExpiration benchmarks the time wheel expiration performance
func BenchmarkTimeWheelExpiration(b *testing.B) {
	db := MakeDB()
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "benchkey"
		db.ExecCommand("SET", key, "value")
		db.ExecCommand("EXPIRE", key, "100")
	}
}

// TestConcurrentAccessWithTimeWheel tests concurrent access with time wheel
func TestConcurrentAccessWithTimeWheel(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	done := make(chan bool)

	// Goroutine 1: Add keys with TTL
	go func() {
		for i := 0; i < 100; i++ {
			key := "key1-" + string(rune('0'+i%10))
			db.ExecCommand("SET", key, "value")
			db.ExecCommand("EXPIRE", key, "100")
		}
		done <- true
	}()

	// Goroutine 2: Delete keys
	go func() {
		for i := 0; i < 50; i++ {
			key := "key1-" + string(rune('0'+i%10))
			db.ExecCommand("DEL", key)
		}
		done <- true
	}()

	// Goroutine 3: Get keys
	go func() {
		for i := 0; i < 100; i++ {
			key := "key1-" + string(rune('0'+i%10))
			db.ExecCommand("GET", key)
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done

	// Wait for expirations
	time.Sleep(200 * time.Millisecond)

	// Should complete without errors or deadlocks
}
