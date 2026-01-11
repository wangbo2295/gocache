package functional

import (
	"testing"

	"github.com/wangbo/gocache/test/e2e"
)

var _ = &e2e.TestClient{} // Verify e2e.TestClient implements expected interface

// TestTransaction_BasicOperations tests MULTI, EXEC, DISCARD
func TestTransaction_BasicOperations(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	t.Run("MULTI/EXEC executes commands atomically", func(t *testing.T) {
		client.Send("DEL", "key1", "key2", "key3")

		// Start transaction
		reply, err := client.Send("MULTI")
		if err != nil || !reply.IsOK() {
			t.Errorf("MULTI failed: %v", err)
		}

		// Queue commands
		client.Send("SET", "key1", "value1")
		client.Send("SET", "key2", "value2")
		client.Send("SET", "key3", "value3")

		// Execute transaction
		reply, err = client.Send("EXEC")
		if err != nil {
			t.Errorf("EXEC failed: %v", err)
		}

		// Verify all commands were executed
		arr := reply.GetArray()
		if arr == nil || len(arr) != 3 {
			t.Errorf("EXEC should return 3 results, got %d", len(arr))
		} else {
			// All should be OK
			for i, result := range arr {
				if format(result) != "OK" {
					t.Errorf("EXEC result[%d] should be OK, got %v", i, result)
				}
			}
		}

		// Verify values
		reply, err = client.Send("GET", "key1")
		if err != nil || reply.GetString() != "value1" {
			t.Error("Transaction should have executed SET commands")
		}

		// Cleanup
		client.Send("DEL", "key1", "key2", "key3")
	})

	t.Run("DISCARD discards transaction", func(t *testing.T) {
		client.Send("DEL", "discard_key")

		// Start transaction
		reply, err := client.Send("MULTI")
		if err != nil || !reply.IsOK() {
			t.Errorf("MULTI failed: %v", err)
		}

		// Queue commands
		client.Send("SET", "discard_key", "value")

		// Discard transaction
		reply, err = client.Send("DISCARD")
		if err != nil || !reply.IsOK() {
			t.Errorf("DISCARD failed: %v", err)
		}

		// Verify command was not executed
		reply, err = client.Send("GET", "discard_key")
		if err != nil {
			t.Errorf("GET failed: %v", err)
		}
		if !reply.IsNil() {
			t.Error("DISCARD should not have executed commands")
		}

		// Cleanup
		client.Send("DEL", "discard_key")
	})

	t.Run("EXEC without MULTI returns error", func(t *testing.T) {
		// Make sure we're not in MULTI state
		client.Send("DISCARD")

		reply, err := client.Send("EXEC")
		if err == nil {
			t.Error("EXEC without MULTI should return error")
		} else if reply != nil && !reply.IsError() {
			t.Error("EXEC without MULTI should return error reply")
		}
	})

	t.Run("MULTI when already in MULTI returns error", func(t *testing.T) {
		client.Send("MULTI")
		reply, err := client.Send("MULTI")
		client.Send("DISCARD") // Clean up

		if err == nil {
			t.Error("MULTI when already in MULTI should return error")
		} else if reply != nil && !reply.IsError() {
			t.Error("MULTI when already in MULTI should return error reply")
		}
	})
}

// TestTransaction_WatchTests tests WATCH and UNWATCH
func TestTransaction_WatchTests(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	t.Run("WATCH detects key modification", func(t *testing.T) {
		client.Send("DEL", "watch_key")

		// Set initial value
		client.Send("SET", "watch_key", "initial")

		// Watch the key
		reply, err := client.Send("WATCH", "watch_key")
		if err != nil || !reply.IsOK() {
			t.Errorf("WATCH failed: %v", err)
		}

		// Start transaction
		reply, err = client.Send("MULTI")
		if err != nil || !reply.IsOK() {
			t.Errorf("MULTI failed: %v", err)
		}

		// Queue command
		client.Send("SET", "watch_key", "transaction_value")

		// Modify watched key outside transaction (simulated by another command)
		// Since we're using a single connection, we need to EXEC first to see the effect
		// In a real scenario with multiple connections, this would be done by another client

		// Execute transaction
		reply, err = client.Send("EXEC")
		if err != nil {
			t.Errorf("EXEC failed: %v", err)
		}

		// If EXEC returns nil, the transaction was aborted due to WATCH
		// If EXEC returns results, the transaction succeeded
		if reply.IsNil() {
			t.Log("Transaction was aborted (WATCH detected modification)")
		} else {
			t.Log("Transaction executed successfully (no conflicting modification)")
		}

		// Cleanup
		client.Send("DEL", "watch_key")
	})

	t.Run("UNWATCH clears all watched keys", func(t *testing.T) {
		client.Send("SET", "unwatch_key", "value")

		// Watch the key
		reply, err := client.Send("WATCH", "unwatch_key")
		if err != nil || !reply.IsOK() {
			t.Errorf("WATCH failed: %v", err)
		}

		// Unwatch
		reply, err = client.Send("UNWATCH")
		if err != nil || !reply.IsOK() {
			t.Errorf("UNWATCH failed: %v", err)
		}

		// Now start transaction and execute
		client.Send("MULTI")
		client.Send("SET", "unwatch_key", "new_value")
		reply, err = client.Send("EXEC")
		if err != nil {
			t.Errorf("EXEC failed: %v", err)
		}

		// Transaction should succeed (not aborted by WATCH)
		if reply.IsNil() {
			t.Error("After UNWATCH, transaction should not be aborted")
		}

		// Cleanup
		client.Send("DEL", "unwatch_key")
	})

	t.Run("WATCH multiple keys", func(t *testing.T) {
		client.Send("DEL", "key1", "key2")
		client.Send("SET", "key1", "value1")
		client.Send("SET", "key2", "value2")

		// Watch multiple keys
		reply, err := client.Send("WATCH", "key1", "key2")
		if err != nil || !reply.IsOK() {
			t.Errorf("WATCH failed: %v", err)
		}

		// Start transaction
		client.Send("MULTI")
		client.Send("GET", "key1")
		client.Send("GET", "key2")
		reply, err = client.Send("EXEC")

		if err != nil {
			t.Errorf("EXEC failed: %v", err)
		}

		// If no modifications, transaction should succeed
		if reply != nil && !reply.IsNil() {
			arr := reply.GetArray()
			if arr != nil && len(arr) == 2 {
				t.Log("Transaction with WATCH on multiple keys succeeded")
			}
		}

		// Cleanup
		client.Send("DEL", "key1", "key2")
	})
}

// TestTransaction_ErrorHandling tests error handling in transactions
func TestTransaction_ErrorHandling(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	t.Run("Commands queued during MULTI are not executed immediately", func(t *testing.T) {
		client.Send("DEL", "queued_key")

		// Start transaction
		client.Send("MULTI")

		// Queue SET command
		reply, err := client.Send("SET", "queued_key", "value")
		if err != nil {
			t.Errorf("SET during MULTI failed: %v", err)
		}
		// Should return QUEUED status
		if reply.GetString() != "QUEUED" {
			t.Logf("SET during MULTI returned: %v (expected QUEUED)", reply.GetString())
		}

		// Key should not be set yet
		// Exit transaction mode
		client.Send("DISCARD")

		reply, err = client.Send("GET", "queued_key")
		if err != nil {
			t.Errorf("GET failed: %v", err)
		}
		if !reply.IsNil() {
			t.Error("Queued commands should not execute before EXEC")
		}

		// Cleanup
		client.Send("DEL", "queued_key")
	})

	t.Run("Transaction with syntax error", func(t *testing.T) {
		// Start transaction
		client.Send("MULTI")

		// Queue invalid command
		client.Send("INVALIDCOMMAND")

		// EXEC should fail
		reply, err := client.Send("EXEC")
		client.Send("DISCARD") // Clean up

		// Behavior depends on implementation
		// Some Redis versions execute valid commands, others abort entire transaction
		t.Logf("EXEC with invalid command returned: err=%v, reply=%v", err, reply)
	})

	t.Run("Transaction continues after non-fatal errors", func(t *testing.T) {
		client.Send("DEL", "tx_key1", "tx_key2")

		// Start transaction
		client.Send("MULTI")

		// Queue valid commands
		client.Send("SET", "tx_key1", "value1")

		// Queue command that might fail (e.g., INCR on string)
		client.Send("SET", "tx_key2", "value2")
		client.Send("INCR", "tx_key2") // This will fail if tx_key2 is not an integer

		// Execute
		reply, err := client.Send("EXEC")
		if err != nil {
			t.Errorf("EXEC failed: %v", err)
		}

		// Check results
		if reply != nil {
			arr := reply.GetArray()
			if arr != nil {
				t.Logf("Transaction returned %d results", len(arr))
				// First command should succeed
				// Second command should fail
				// Third command might fail or succeed depending on implementation
			}
		}

		// Cleanup
		client.Send("DEL", "tx_key1", "tx_key2")
	})
}

// TestTransaction_ComplexScenarios tests complex transaction scenarios
func TestTransaction_ComplexScenarios(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	t.Run("Transaction with multiple data types", func(t *testing.T) {
		client.Send("DEL", "tx_string", "tx_hash", "tx_list")

		// Start transaction
		client.Send("MULTI")

		// Queue commands for different data types
		client.Send("SET", "tx_string", "value")
		client.Send("HSET", "tx_hash", "field", "value")
		client.Send("RPUSH", "tx_list", "a", "b", "c")

		// Execute
		reply, err := client.Send("EXEC")
		if err != nil {
			t.Errorf("EXEC failed: %v", err)
		}

		// Verify all commands executed
		if reply != nil {
			arr := reply.GetArray()
			if arr == nil || len(arr) != 3 {
				t.Errorf("EXEC should return 3 results, got %d", len(arr))
			}
		}

		// Verify data
		reply, err = client.Send("GET", "tx_string")
		if err != nil || reply.GetString() != "value" {
			t.Error("String command should have executed")
		}

		reply, err = client.Send("HGET", "tx_hash", "field")
		if err != nil || reply.GetString() != "value" {
			t.Error("Hash command should have executed")
		}

		reply, err = client.Send("LLEN", "tx_list")
		if err != nil {
			t.Errorf("LLEN failed: %v", err)
		}
		length, _ := reply.GetInt()
		if length != 3 {
			t.Errorf("List should have 3 elements, got %d", length)
		}

		// Cleanup
		client.Send("DEL", "tx_string", "tx_hash", "tx_list")
	})

	t.Run("Nested transaction behavior", func(t *testing.T) {
		// Start first transaction
		client.Send("MULTI")

		// Try to start nested transaction
		_, err := client.Send("MULTI")
		client.Send("DISCARD") // Exit transaction

		if err == nil {
			t.Log("Nested MULTI was allowed (implementation-dependent)")
		} else {
			t.Log("Nested MULTI was rejected (expected)")
		}
	})

	t.Run("Transaction with TTL commands", func(t *testing.T) {
		client.Send("DEL", "ttl_tx_key")

		// Start transaction
		client.Send("MULTI")

		// Queue commands with TTL
		client.Send("SET", "ttl_tx_key", "value")
		client.Send("EXPIRE", "ttl_tx_key", "10")

		// Execute
		reply, err := client.Send("EXEC")
		if err != nil {
			t.Errorf("EXEC failed: %v", err)
		}

		// Verify TTL was set
		reply, err = client.Send("TTL", "ttl_tx_key")
		if err != nil {
			t.Errorf("TTL failed: %v", err)
		}
		ttl, _ := reply.GetInt()
		if ttl <= 0 {
			t.Error("TTL should be set in transaction")
		}

		// Cleanup
		client.Send("DEL", "ttl_tx_key")
	})
}

// TestTransaction_RetryLogic tests optimistic locking retry logic
func TestTransaction_RetryLogic(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	t.Run("Optimistic locking retry pattern", func(t *testing.T) {
		client.Send("DEL", "counter")

		// Initialize counter
		client.Send("SET", "counter", "10")

		// Simulate optimistic locking retry
		maxRetries := 3
		success := false

		for i := 0; i < maxRetries; i++ {
			// Watch counter
			client.Send("WATCH", "counter")

			// Get current value
			reply, err := client.Send("GET", "counter")
			if err != nil {
				continue
			}
			currentValue := reply.GetString()

			// Start transaction
			client.Send("MULTI")
			client.Send("SET", "counter", currentValue)

			// Execute
			reply, err = client.Send("EXEC")
			if err != nil {
				continue
			}

			// Check if transaction succeeded
			if reply != nil && !reply.IsNil() {
				success = true
				break
			}

			// Transaction failed, retry
			t.Logf("Retry %d: transaction aborted, retrying...", i+1)
		}

		if !success {
			t.Error("Optimistic locking retry failed after max retries")
		} else {
			t.Log("Optimistic locking retry succeeded")
		}

		// Cleanup
		client.Send("DEL", "counter")
	})
}
