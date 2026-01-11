package functional

import (
	"testing"
	"time"

	"github.com/wangbo/gocache/test/e2e"
)

var _ = &e2e.TestClient{} // Verify e2e.TestClient implements expected interface

// TestTTL_BasicOperations tests EXPIRE, TTL, PERSIST commands
func TestTTL_BasicOperations(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	t.Run("EXPIRE sets TTL on key", func(t *testing.T) {
		client.Send("DEL", "ttl_key")
		client.Send("SET", "ttl_key", "value")

		reply, err := client.Send("EXPIRE", "ttl_key", "10")
		if err != nil {
			t.Errorf("EXPIRE failed: %v", err)
		}
		result, _ := reply.GetInt()
		if result != 1 {
			t.Errorf("EXPIRE should return 1, got %d", result)
		}

		// Check TTL
		reply, err = client.Send("TTL", "ttl_key")
		if err != nil {
			t.Errorf("TTL failed: %v", err)
		}
		ttl, _ := reply.GetInt()
		if ttl < 1 || ttl > 10 {
			t.Errorf("TTL should be between 1 and 10, got %d", ttl)
		}
	})

	t.Run("EXPIRE on non-existent key returns 0", func(t *testing.T) {
		client.Send("DEL", "nonexistent")
		reply, err := client.Send("EXPIRE", "nonexistent", "10")
		if err != nil {
			t.Errorf("EXPIRE failed: %v", err)
		}
		result, _ := reply.GetInt()
		if result != 0 {
			t.Errorf("EXPIRE on non-existent key should return 0, got %d", result)
		}
	})

	t.Run("TTL on non-existent key returns -2", func(t *testing.T) {
		client.Send("DEL", "nonexistent")
		reply, err := client.Send("TTL", "nonexistent")
		if err != nil {
			t.Errorf("TTL failed: %v", err)
		}
		ttl, _ := reply.GetInt()
		if ttl != -2 {
			t.Errorf("TTL on non-existent key should return -2, got %d", ttl)
		}
	})

	t.Run("TTL on key without expiration returns -1", func(t *testing.T) {
		client.Send("SET", "no_expire_key", "value")
		reply, err := client.Send("TTL", "no_expire_key")
		if err != nil {
			t.Errorf("TTL failed: %v", err)
		}
		ttl, _ := reply.GetInt()
		if ttl != -1 {
			t.Errorf("TTL on key without expiration should return -1, got %d", ttl)
		}
		client.Send("DEL", "no_expire_key")
	})

	t.Run("PERSIST removes TTL from key", func(t *testing.T) {
		client.Send("SET", "persist_key", "value")
		client.Send("EXPIRE", "persist_key", "10")

		reply, err := client.Send("PERSIST", "persist_key")
		if err != nil {
			t.Errorf("PERSIST failed: %v", err)
		}
		result, _ := reply.GetInt()
		if result != 1 {
			t.Errorf("PERSIST should return 1, got %d", result)
		}

		// Verify TTL is -1 (no expiration)
		reply, err = client.Send("TTL", "persist_key")
		if err != nil {
			t.Errorf("TTL failed: %v", err)
		}
		ttl, _ := reply.GetInt()
		if ttl != -1 {
			t.Errorf("After PERSIST, TTL should be -1, got %d", ttl)
		}
		client.Send("DEL", "persist_key")
	})

	t.Run("PERSIST on key without TTL returns 0", func(t *testing.T) {
		client.Send("SET", "no_ttl_key", "value")
		reply, err := client.Send("PERSIST", "no_ttl_key")
		if err != nil {
			t.Errorf("PERSIST failed: %v", err)
		}
		result, _ := reply.GetInt()
		if result != 0 {
			t.Errorf("PERSIST on key without TTL should return 0, got %d", result)
		}
		client.Send("DEL", "no_ttl_key")
	})

	// Cleanup
	client.Send("DEL", "ttl_key")
}

// TestTTL_PrecisionOperations tests PEXPIRE, PTTL commands
func TestTTL_PrecisionOperations(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	t.Run("PEXPIRE sets TTL in milliseconds", func(t *testing.T) {
		client.Send("SET", "pttl_key", "value")
		reply, err := client.Send("PEXPIRE", "pttl_key", "5000")
		if err != nil {
			t.Errorf("PEXPIRE failed: %v", err)
		}
		result, _ := reply.GetInt()
		if result != 1 {
			t.Errorf("PEXPIRE should return 1, got %d", result)
		}

		// Check PTTL
		reply, err = client.Send("PTTL", "pttl_key")
		if err != nil {
			t.Errorf("PTTL failed: %v", err)
		}
		pttl, _ := reply.GetInt()
		if pttl < 1000 || pttl > 5000 {
			t.Errorf("PTTL should be between 1000 and 5000, got %d", pttl)
		}
		client.Send("DEL", "pttl_key")
	})

	t.Run("PTTL on non-existent key returns -2", func(t *testing.T) {
		client.Send("DEL", "nonexistent")
		reply, err := client.Send("PTTL", "nonexistent")
		if err != nil {
			t.Errorf("PTTL failed: %v", err)
		}
		pttl, _ := reply.GetInt()
		if pttl != -2 {
			t.Errorf("PTTL on non-existent key should return -2, got %d", pttl)
		}
	})

	t.Run("PTTL on key without expiration returns -1", func(t *testing.T) {
		client.Send("SET", "no_expire", "value")
		reply, err := client.Send("PTTL", "no_expire")
		if err != nil {
			t.Errorf("PTTL failed: %v", err)
		}
		pttl, _ := reply.GetInt()
		if pttl != -1 {
			t.Errorf("PTTL on key without expiration should return -1, got %d", pttl)
		}
		client.Send("DEL", "no_expire")
	})
}

// TestTTL_Expiration tests key expiration behavior
func TestTTL_Expiration(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	t.Run("Key expires after TTL", func(t *testing.T) {
		client.Send("SET", "expire_test", "value")
		client.Send("EXPIRE", "expire_test", "1")

		// Key should exist initially
		reply, err := client.Send("GET", "expire_test")
		if err != nil {
			t.Errorf("GET failed: %v", err)
		}
		if reply.GetString() != "value" {
			t.Error("Key should exist before expiration")
		}

		// Wait for expiration
		time.Sleep(2 * time.Second)

		// Key should be expired
		reply, err = client.Send("GET", "expire_test")
		if err != nil {
			t.Errorf("GET failed: %v", err)
		}
		if !reply.IsNil() {
			t.Error("Key should be expired after TTL")
		}

		// TTL should return -2
		reply, err = client.Send("TTL", "expire_test")
		if err != nil {
			t.Errorf("TTL failed: %v", err)
		}
		ttl, _ := reply.GetInt()
		if ttl != -2 {
			t.Errorf("TTL on expired key should return -2, got %d", ttl)
		}
	})

	t.Run("Key with millisecond TTL expires", func(t *testing.T) {
		client.Send("SET", "pexpire_test", "value")
		client.Send("PEXPIRE", "pexpire_test", "500")

		// Wait for expiration
		time.Sleep(1 * time.Second)

		// Key should be expired
		reply, err := client.Send("GET", "pexpire_test")
		if err != nil {
			t.Errorf("GET failed: %v", err)
		}
		if !reply.IsNil() {
			t.Error("Key should be expired after PTTL")
		}
	})
}

// TestTTL_Overwrite tests TTL behavior when key is overwritten
func TestTTL_Overwrite(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	t.Run("SET overwrites TTL", func(t *testing.T) {
		client.Send("SET", "overwrite_key", "value1")
		client.Send("EXPIRE", "overwrite_key", "10")

		// Verify TTL is set
		reply, err := client.Send("TTL", "overwrite_key")
		if err != nil {
			t.Errorf("TTL failed: %v", err)
		}
		ttl, _ := reply.GetInt()
		if ttl <= 0 {
			t.Error("TTL should be set")
		}

		// Overwrite key without TTL options
		client.Send("SET", "overwrite_key", "value2")

		// TTL should be removed
		reply, err = client.Send("TTL", "overwrite_key")
		if err != nil {
			t.Errorf("TTL failed: %v", err)
		}
		ttl, _ = reply.GetInt()
		if ttl != -1 {
			t.Errorf("After SET overwrite, TTL should be -1, got %d", ttl)
		}

		client.Send("DEL", "overwrite_key")
	})
}

// TestTTL_AllDataTypes tests TTL on all data types
func TestTTL_AllDataTypes(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	t.Run("TTL works on String", func(t *testing.T) {
		client.Send("SET", "string_ttl", "value")
		client.Send("EXPIRE", "string_ttl", "5")
		reply, err := client.Send("TTL", "string_ttl")
		if err != nil {
			t.Errorf("TTL failed: %v", err)
		}
		ttl, _ := reply.GetInt()
		if ttl <= 0 {
			t.Error("TTL should be set on String")
		}
		client.Send("DEL", "string_ttl")
	})

	t.Run("TTL works on Hash", func(t *testing.T) {
		client.Send("HSET", "hash_ttl", "field", "value")
		client.Send("EXPIRE", "hash_ttl", "5")
		reply, err := client.Send("TTL", "hash_ttl")
		if err != nil {
			t.Errorf("TTL failed: %v", err)
		}
		ttl, _ := reply.GetInt()
		if ttl <= 0 {
			t.Error("TTL should be set on Hash")
		}
		client.Send("DEL", "hash_ttl")
	})

	t.Run("TTL works on List", func(t *testing.T) {
		client.Send("RPUSH", "list_ttl", "a", "b", "c")
		client.Send("EXPIRE", "list_ttl", "5")
		reply, err := client.Send("TTL", "list_ttl")
		if err != nil {
			t.Errorf("TTL failed: %v", err)
		}
		ttl, _ := reply.GetInt()
		if ttl <= 0 {
			t.Error("TTL should be set on List")
		}
		client.Send("DEL", "list_ttl")
	})

	t.Run("TTL works on Set", func(t *testing.T) {
		client.Send("SADD", "set_ttl", "a", "b", "c")
		client.Send("EXPIRE", "set_ttl", "5")
		reply, err := client.Send("TTL", "set_ttl")
		if err != nil {
			t.Errorf("TTL failed: %v", err)
		}
		ttl, _ := reply.GetInt()
		if ttl <= 0 {
			t.Error("TTL should be set on Set")
		}
		client.Send("DEL", "set_ttl")
	})

	t.Run("TTL works on SortedSet", func(t *testing.T) {
		client.Send("ZADD", "zset_ttl", "1", "one")
		client.Send("EXPIRE", "zset_ttl", "5")
		reply, err := client.Send("TTL", "zset_ttl")
		if err != nil {
			t.Errorf("TTL failed: %v", err)
		}
		ttl, _ := reply.GetInt()
		if ttl <= 0 {
			t.Error("TTL should be set on SortedSet")
		}
		client.Send("DEL", "zset_ttl")
	})
}
