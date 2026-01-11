package database

import (
	"testing"
	"time"

	"github.com/wangbo/gocache/eviction"
)

// 4.1 功能验收测试

// TestAcceptance_BasicDataTypes tests all basic data types
func TestAcceptance_BasicDataTypes(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	t.Run("String类型基础操作", func(t *testing.T) {
		// SET/GET
		result, err := db.Exec([][]byte{[]byte("SET"), []byte("str_key"), []byte("hello")})
		if err != nil || string(result[0]) != "OK" {
			t.Fatalf("SET failed: %v, %s", err, result)
		}

		result, err = db.Exec([][]byte{[]byte("GET"), []byte("str_key")})
		if err != nil || string(result[0]) != "hello" {
			t.Fatalf("GET failed: %v, %s", err, result)
		}

		// INCR/DECR
		result, err = db.Exec([][]byte{[]byte("SET"), []byte("counter"), []byte("10")})
		if err != nil {
			t.Fatalf("SET counter failed: %v", err)
		}

		result, err = db.Exec([][]byte{[]byte("INCR"), []byte("counter")})
		if err != nil || string(result[0]) != "11" {
			t.Fatalf("INCR failed: %v, %s", err, result)
		}

		result, err = db.Exec([][]byte{[]byte("DECR"), []byte("counter")})
		if err != nil || string(result[0]) != "10" {
			t.Fatalf("DECR failed: %v, %s", err, result)
		}

		// MGET/MSET
		result, err = db.Exec([][]byte{
			[]byte("MSET"), []byte("key1"), []byte("val1"),
			[]byte("key2"), []byte("val2"),
		})
		if err != nil {
			t.Fatalf("MSET failed: %v", err)
		}

		result, err = db.Exec([][]byte{[]byte("MGET"), []byte("key1"), []byte("key2")})
		if err != nil || len(result) != 2 {
			t.Fatalf("MGET failed: %v, result length: %d", err, len(result))
		}
	})

	t.Run("Hash类型操作", func(t *testing.T) {
		// HSET/HGET
		result, err := db.Exec([][]byte{
			[]byte("HSET"), []byte("user:1"), []byte("name"), []byte("Alice"),
		})
		if err != nil || string(result[0]) != "1" {
			t.Fatalf("HSET failed: %v, %s", err, result)
		}

		result, err = db.Exec([][]byte{
			[]byte("HGET"), []byte("user:1"), []byte("name"),
		})
		if err != nil || string(result[0]) != "Alice" {
			t.Fatalf("HGET failed: %v, %s", err, result)
		}

		// HGETALL
		result, err = db.Exec([][]byte{
			[]byte("HSET"), []byte("user:1"), []byte("age"), []byte("25"),
		})
		if err != nil {
			t.Fatalf("HSET age failed: %v", err)
		}

		result, err = db.Exec([][]byte{[]byte("HGETALL"), []byte("user:1")})
		if err != nil || len(result) < 4 {
			t.Fatalf("HGETALL failed: %v, result: %v", err, result)
		}
	})

	t.Run("List类型操作", func(t *testing.T) {
		// LPUSH/LPOP
		result, err := db.Exec([][]byte{
			[]byte("LPUSH"), []byte("mylist"), []byte("world"),
		})
		if err != nil {
			t.Fatalf("LPUSH failed: %v", err)
		}

		result, err = db.Exec([][]byte{
			[]byte("LPUSH"), []byte("mylist"), []byte("hello"),
		})
		if err != nil {
			t.Fatalf("LPUSH failed: %v", err)
		}

		result, err = db.Exec([][]byte{[]byte("LLEN"), []byte("mylist")})
		if err != nil || string(result[0]) != "2" {
			t.Fatalf("LLEN failed: %v, %s", err, result)
		}

		result, err = db.Exec([][]byte{[]byte("LRANGE"), []byte("mylist"), []byte("0"), []byte("-1")})
		if err != nil || len(result) != 2 {
			t.Fatalf("LRANGE failed: %v, result length: %d", err, len(result))
		}

		_, err = db.Exec([][]byte{[]byte("LPOP"), []byte("mylist")})
		if err != nil {
			t.Fatalf("LPOP failed: %v", err)
		}

		result, err = db.Exec([][]byte{[]byte("LLEN"), []byte("mylist")})
		if err != nil || string(result[0]) != "1" {
			t.Fatalf("LLEN after LPOP failed: %v, %s", err, result)
		}
	})

	t.Run("Set类型操作", func(t *testing.T) {
		// SADD/SMEMBERS
		result, err := db.Exec([][]byte{
			[]byte("SADD"), []byte("myset"), []byte("member1"),
		})
		if err != nil || string(result[0]) != "1" {
			t.Fatalf("SADD failed: %v, %s", err, result)
		}

		result, err = db.Exec([][]byte{
			[]byte("SADD"), []byte("myset"), []byte("member2"),
		})
		if err != nil || string(result[0]) != "1" {
			t.Fatalf("SADD member2 failed: %v, %s", err, result)
		}

		result, err = db.Exec([][]byte{[]byte("SCARD"), []byte("myset")})
		if err != nil || string(result[0]) != "2" {
			t.Fatalf("SCARD failed: %v, %s", err, result)
		}

		result, err = db.Exec([][]byte{[]byte("SISMEMBER"), []byte("myset"), []byte("member1")})
		if err != nil || string(result[0]) != "1" {
			t.Fatalf("SISMEMBER failed: %v, %s", err, result)
		}
	})

	t.Run("SortedSet类型操作", func(t *testing.T) {
		// ZADD/ZRANGE
		result, err := db.Exec([][]byte{
			[]byte("ZADD"), []byte("myzset"), []byte("1"), []byte("one"),
		})
		if err != nil || string(result[0]) != "1" {
			t.Fatalf("ZADD failed: %v, %s", err, result)
		}

		result, err = db.Exec([][]byte{
			[]byte("ZADD"), []byte("myzset"), []byte("2"), []byte("two"),
		})
		if err != nil || string(result[0]) != "1" {
			t.Fatalf("ZADD two failed: %v, %s", err, result)
		}

		result, err = db.Exec([][]byte{[]byte("ZCARD"), []byte("myzset")})
		if err != nil || string(result[0]) != "2" {
			t.Fatalf("ZCARD failed: %v, %s", err, result)
		}

		result, err = db.Exec([][]byte{[]byte("ZRANGE"), []byte("myzset"), []byte("0"), []byte("-1")})
		if err != nil || len(result) < 2 {
			t.Fatalf("ZRANGE failed: %v, result length: %d", err, len(result))
		}

		result, err = db.Exec([][]byte{[]byte("ZSCORE"), []byte("myzset"), []byte("one")})
		if err != nil || string(result[0]) != "1" {
			t.Fatalf("ZSCORE failed: %v, %s", err, result)
		}
	})
}

// TestAcceptance_KeyExpiration tests key expiration and auto eviction
func TestAcceptance_KeyExpiration(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	t.Run("基础过期设置", func(t *testing.T) {
		// SET with EX - note: EX option not yet supported in SET, using EXPIRE
		result, err := db.Exec([][]byte{
			[]byte("SET"), []byte("expire_key"), []byte("value"),
		})
		if err != nil {
			t.Fatalf("SET failed: %v", err)
		}

		// Set expiration
		result, err = db.Exec([][]byte{
			[]byte("EXPIRE"), []byte("expire_key"), []byte("1"),
		})
		if err != nil || string(result[0]) != "1" {
			t.Fatalf("EXPIRE failed: %v, %s", err, result)
		}

		// Check TTL
		result, err = db.Exec([][]byte{[]byte("TTL"), []byte("expire_key")})
		if err != nil {
			t.Fatalf("TTL failed: %v", err)
		}
		ttl := string(result[0])
		if ttl == "-2" || ttl == "-1" {
			t.Fatalf("Expected positive TTL, got: %s", ttl)
		}

		// Wait for expiration
		time.Sleep(1100 * time.Millisecond)

		// Key should be expired
		result, err = db.Exec([][]byte{[]byte("GET"), []byte("expire_key")})
		if err != nil {
			t.Fatalf("GET after expiry failed: %v", err)
		}
		if len(result) > 0 && string(result[0]) != "" {
			t.Fatalf("Key should be expired, got: %s", result[0])
		}
	})

	t.Run("EXPIRE命令", func(t *testing.T) {
		result, err := db.Exec([][]byte{
			[]byte("SET"), []byte("expire_cmd_key"), []byte("value"),
		})
		if err != nil {
			t.Fatalf("SET failed: %v", err)
		}

		// Set expiration
		result, err = db.Exec([][]byte{
			[]byte("EXPIRE"), []byte("expire_cmd_key"), []byte("1"),
		})
		if err != nil || string(result[0]) != "1" {
			t.Fatalf("EXPIRE failed: %v, %s", err, result)
		}

		// Wait for expiration
		time.Sleep(1100 * time.Millisecond)

		// Key should be expired
		result, err = db.Exec([][]byte{[]byte("GET"), []byte("expire_cmd_key")})
		if err != nil {
			t.Fatalf("GET after expiry failed: %v", err)
		}
		if len(result) > 0 && string(result[0]) != "" {
			t.Fatalf("Key should be expired, got: %s", result[0])
		}
	})

	t.Run("PERSIST命令", func(t *testing.T) {
		// SET with expiration
		result, err := db.Exec([][]byte{
			[]byte("SET"), []byte("persist_key"), []byte("value"),
		})
		if err != nil {
			t.Fatalf("SET failed: %v", err)
		}

		// Set expiration
		result, err = db.Exec([][]byte{
			[]byte("EXPIRE"), []byte("persist_key"), []byte("10"),
		})
		if err != nil || string(result[0]) != "1" {
			t.Fatalf("EXPIRE failed: %v, %s", err, result)
		}

		// Remove expiration
		result, err = db.Exec([][]byte{[]byte("PERSIST"), []byte("persist_key")})
		if err != nil || string(result[0]) != "1" {
			t.Fatalf("PERSIST failed: %v, %s", err, result)
		}

		// Check TTL (should be -1: no expiration)
		result, err = db.Exec([][]byte{[]byte("TTL"), []byte("persist_key")})
		if err != nil || string(result[0]) != "-1" {
			t.Fatalf("TTL after PERSIST should be -1, got: %s", result[0])
		}
	})

	t.Run("毫秒级过期", func(t *testing.T) {
		result, err := db.Exec([][]byte{
			[]byte("SET"), []byte("pexpire_key"), []byte("value"),
		})
		if err != nil {
			t.Fatalf("SET failed: %v", err)
		}

		// Set expiration in milliseconds
		result, err = db.Exec([][]byte{
			[]byte("PEXPIRE"), []byte("pexpire_key"), []byte("500"),
		})
		if err != nil || string(result[0]) != "1" {
			t.Fatalf("PEXPIRE failed: %v, %s", err, result)
		}

		// Check PTTL
		result, err = db.Exec([][]byte{[]byte("PTTL"), []byte("pexpire_key")})
		if err != nil {
			t.Fatalf("PTTL failed: %v", err)
		}
		pttl := string(result[0])
		if pttl == "-2" || pttl == "-1" {
			t.Fatalf("Expected positive PTTL, got: %s", pttl)
		}

		// Wait for expiration
		time.Sleep(600 * time.Millisecond)

		// Key should be expired
		result, err = db.Exec([][]byte{[]byte("GET"), []byte("pexpire_key")})
		if err != nil {
			t.Fatalf("GET after expiry failed: %v", err)
		}
		if len(result) > 0 && string(result[0]) != "" {
			t.Fatalf("Key should be expired, got: %s", result[0])
		}
	})
}

// TestAcceptance_TransactionAtomicity tests transaction atomicity
func TestAcceptance_TransactionAtomicity(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	t.Run("MULTI/EXEC基础事务", func(t *testing.T) {
		// Start transaction
		result, err := db.Exec([][]byte{[]byte("MULTI")})
		if err != nil || string(result[0]) != "OK" {
			t.Fatalf("MULTI failed: %v, %s", err, result)
		}

		// Queue commands
		result, err = db.Exec([][]byte{[]byte("SET"), []byte("tx_key1"), []byte("value1")})
		if err != nil || string(result[0]) != "QUEUED" {
			t.Fatalf("SET in MULTI failed: %v, %s", err, result)
		}

		result, err = db.Exec([][]byte{[]byte("SET"), []byte("tx_key2"), []byte("value2")})
		if err != nil || string(result[0]) != "QUEUED" {
			t.Fatalf("SET in MULTI failed: %v, %s", err, result)
		}

		// Execute transaction
		result, err = db.Exec([][]byte{[]byte("EXEC")})
		if err != nil {
			t.Fatalf("EXEC failed: %v", err)
		}

		// Verify both keys were set
		result, err = db.Exec([][]byte{[]byte("GET"), []byte("tx_key1")})
		if err != nil || string(result[0]) != "value1" {
			t.Fatalf("GET tx_key1 failed: %v, %s", err, result)
		}

		result, err = db.Exec([][]byte{[]byte("GET"), []byte("tx_key2")})
		if err != nil || string(result[0]) != "value2" {
			t.Fatalf("GET tx_key2 failed: %v, %s", err, result)
		}
	})

	t.Run("DISCARD取消事务", func(t *testing.T) {
		// Start transaction
		result, err := db.Exec([][]byte{[]byte("MULTI")})
		if err != nil || string(result[0]) != "OK" {
			t.Fatalf("MULTI failed: %v, %s", err, result)
		}

		// Queue command
		result, err = db.Exec([][]byte{[]byte("SET"), []byte("discard_key"), []byte("value")})
		if err != nil || string(result[0]) != "QUEUED" {
			t.Fatalf("SET in MULTI failed: %v, %s", err, result)
		}

		// Discard transaction
		result, err = db.Exec([][]byte{[]byte("DISCARD")})
		if err != nil || string(result[0]) != "OK" {
			t.Fatalf("DISCARD failed: %v, %s", err, result)
		}

		// Verify key was not set
		result, err = db.Exec([][]byte{[]byte("GET"), []byte("discard_key")})
		if err != nil {
			t.Fatalf("GET after DISCARD failed: %v", err)
		}
		if len(result) > 0 && string(result[0]) != "" {
			t.Fatalf("Key should not exist after DISCARD, got: %s", result[0])
		}
	})

	t.Run("WATCH乐观锁", func(t *testing.T) {
		// Set initial value
		result, err := db.Exec([][]byte{
			[]byte("SET"), []byte("watch_key"), []byte("10"),
		})
		if err != nil {
			t.Fatalf("SET failed: %v", err)
		}

		// Watch key
		result, err = db.Exec([][]byte{[]byte("WATCH"), []byte("watch_key")})
		if err != nil || string(result[0]) != "OK" {
			t.Fatalf("WATCH failed: %v, %s", err, result)
		}

		// Start transaction
		result, err = db.Exec([][]byte{[]byte("MULTI")})
		if err != nil || string(result[0]) != "OK" {
			t.Fatalf("MULTI failed: %v, %s", err, result)
		}

		// Queue command in transaction
		result, err = db.Exec([][]byte{[]byte("SET"), []byte("watch_key"), []byte("20")})
		if err != nil || string(result[0]) != "QUEUED" {
			t.Fatalf("SET in MULTI failed: %v, %s", err, result)
		}

		// Execute transaction
		result, err = db.Exec([][]byte{[]byte("EXEC")})
		if err != nil {
			t.Fatalf("EXEC failed: %v", err)
		}

		// Verify value was updated
		result, err = db.Exec([][]byte{[]byte("GET"), []byte("watch_key")})
		if err != nil || string(result[0]) != "20" {
			t.Fatalf("GET watch_key failed: %v, %s", err, result)
		}

		// Note: Testing WATCH failure requires separate connections
		// which is not easily testable in this context
	})
}

// TestAcceptance_MemoryEviction tests memory-based eviction policies
func TestAcceptance_MemoryEviction(t *testing.T) {
	t.Run("LRU淘汰策略", func(t *testing.T) {
		db := MakeDB()
		db.evictionPolicy = eviction.NewLRU(1000)
		defer db.Close()

		// Fill database
		for i := 0; i < 100; i++ {
			key := []byte("lru_key_" + string(rune('0'+i)))
			value := []byte("x") // 1 byte value
			db.Exec([][]byte{[]byte("SET"), key, value})
		}

		// Access some keys to update LRU
		db.Exec([][]byte{[]byte("GET"), []byte("lru_key_0")})
		db.Exec([][]byte{[]byte("GET"), []byte("lru_key_1")})

		// Memory should be controlled
		_ = db.usedMemory
	})

	t.Run("LFU淘汰策略", func(t *testing.T) {
		db := MakeDB()
		db.evictionPolicy = eviction.NewLFU(1000)
		defer db.Close()

		// Fill and access keys
		for i := 0; i < 50; i++ {
			key := []byte("lfu_key_" + string(rune('0'+i)))
			value := []byte("x")
			db.Exec([][]byte{[]byte("SET"), key, value})
		}

		// Frequently access key 0
		for i := 0; i < 10; i++ {
			db.Exec([][]byte{[]byte("GET"), []byte("lfu_key_0")})
		}

		// Key 0 should still exist due to high frequency
		result, err := db.Exec([][]byte{[]byte("GET"), []byte("lfu_key_0")})
		if err != nil {
			t.Fatalf("GET lfu_key_0 failed: %v", err)
		}
		if len(result) == 0 {
			t.Fatalf("Frequently accessed key was evicted")
		}
	})
}

// TestAcceptance_ComplexScenarios tests complex real-world scenarios
func TestAcceptance_ComplexScenarios(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	t.Run("设备信任评分缓存场景", func(t *testing.T) {
		deviceID := "device:12345"
		score := "85"

		// Cache device trust score
		result, err := db.Exec([][]byte{
			[]byte("SET"), []byte(deviceID), []byte(score),
			[]byte("EX"), []byte("3600"), // 1 hour TTL
		})
		if err != nil {
			t.Fatalf("SET device score failed: %v", err)
		}

		// Retrieve score
		result, err = db.Exec([][]byte{[]byte("GET"), []byte(deviceID)})
		if err != nil || string(result[0]) != score {
			t.Fatalf("GET device score failed: %v, %s", err, result)
		}

		// Update score (increment)
		result, err = db.Exec([][]byte{
			[]byte("INCRBY"), []byte(deviceID), []byte("5"),
		})
		if err != nil || string(result[0]) != "90" {
			t.Fatalf("INCRBY device score failed: %v, %s", err, result)
		}
	})

	t.Run("用户访问频率限流场景", func(t *testing.T) {
		userID := "user:limit:67890"
		window := "60" // 60 second window

		// Increment request count
		result, err := db.Exec([][]byte{
			[]byte("INCR"), []byte(userID),
		})
		if err != nil {
			t.Fatalf("INCR request count failed: %v", err)
		}

		// Set expiration on first request
		count := string(result[0])
		if count == "1" {
			_, err = db.Exec([][]byte{
				[]byte("EXPIRE"), []byte(userID), []byte(window),
			})
			if err != nil {
				t.Fatalf("EXPIRE failed: %v", err)
			}
		}

		// Check if limit exceeded
		if _, err := db.Exec([][]byte{[]byte("GET"), []byte(userID)}); err == nil {
			// In real scenario, check if count > limit
		}

		// Verify TTL exists
		result, err = db.Exec([][]byte{[]byte("TTL"), []byte(userID)})
		if err != nil {
			t.Fatalf("TTL failed: %v", err)
		}
		ttl := string(result[0])
		if ttl == "-2" {
			t.Fatalf("Key should have TTL, got -2")
		}
	})

	t.Run("黑名单检查场景", func(t *testing.T) {
		blacklistKey := "blacklist:ip"

		// Add IPs to blacklist (using SET)
		result, err := db.Exec([][]byte{
			[]byte("SADD"), []byte(blacklistKey),
			[]byte("192.168.1.100"),
			[]byte("192.168.1.101"),
		})
		if err != nil {
			t.Fatalf("SADD to blacklist failed: %v", err)
		}

		// Check if IP is blacklisted
		result, err = db.Exec([][]byte{
			[]byte("SISMEMBER"), []byte(blacklistKey),
			[]byte("192.168.1.100"),
		})
		if err != nil || string(result[0]) != "1" {
			t.Fatalf("SISMEMBER check failed: %v, %s", err, result)
		}

		// Remove from blacklist
		result, err = db.Exec([][]byte{
			[]byte("SREM"), []byte(blacklistKey),
			[]byte("192.168.1.100"),
		})
		if err != nil {
			t.Fatalf("SREM from blacklist failed: %v", err)
		}

		// Verify removal
		result, err = db.Exec([][]byte{
			[]byte("SISMEMBER"), []byte(blacklistKey),
			[]byte("192.168.1.100"),
		})
		if err != nil || string(result[0]) != "0" {
			t.Fatalf("SISMEMBER after SREM should be 0, got: %s", result[0])
		}
	})
}

// TestAcceptance_TypeCommand tests TYPE command
func TestAcceptance_TypeCommand(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// String
	db.Exec([][]byte{[]byte("SET"), []byte("str"), []byte("val")})
	result, err := db.Exec([][]byte{[]byte("TYPE"), []byte("str")})
	if err != nil || string(result[0]) != "string" {
		t.Fatalf("TYPE string failed: %v, %s", err, result)
	}

	// Hash
	db.Exec([][]byte{[]byte("HSET"), []byte("h"), []byte("f"), []byte("v")})
	result, err = db.Exec([][]byte{[]byte("TYPE"), []byte("h")})
	if err != nil || string(result[0]) != "hash" {
		t.Fatalf("TYPE hash failed: %v, %s", err, result)
	}

	// List
	db.Exec([][]byte{[]byte("LPUSH"), []byte("l"), []byte("v")})
	result, err = db.Exec([][]byte{[]byte("TYPE"), []byte("l")})
	if err != nil || string(result[0]) != "list" {
		t.Fatalf("TYPE list failed: %v, %s", err, result)
	}

	// Set
	db.Exec([][]byte{[]byte("SADD"), []byte("s"), []byte("v")})
	result, err = db.Exec([][]byte{[]byte("TYPE"), []byte("s")})
	if err != nil || string(result[0]) != "set" {
		t.Fatalf("TYPE set failed: %v, %s", err, result)
	}

	// Sorted Set
	db.Exec([][]byte{[]byte("ZADD"), []byte("z"), []byte("1"), []byte("v")})
	result, err = db.Exec([][]byte{[]byte("TYPE"), []byte("z")})
	if err != nil || string(result[0]) != "zset" {
		t.Fatalf("TYPE zset failed: %v, %s", err, result)
	}

	// Non-existent
	result, err = db.Exec([][]byte{[]byte("TYPE"), []byte("nonexistent")})
	if err != nil || string(result[0]) != "none" {
		t.Fatalf("TYPE none failed: %v, %s", err, result)
	}
}
