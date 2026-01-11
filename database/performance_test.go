package database

import (
	"testing"
	"time"
)

// BenchmarkSET 性能测试 SET 命令
func BenchmarkSET(b *testing.B) {
	db := MakeDB()
	cmdLine := [][]byte{[]byte("SET"), []byte("key"), []byte("value")}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		db.Exec(cmdLine)
	}
}

// BenchmarkGET 性能测试 GET 命令
func BenchmarkGET(b *testing.B) {
	db := MakeDB()
	db.Exec([][]byte{[]byte("SET"), []byte("key"), []byte("value")})

	cmdLine := [][]byte{[]byte("GET"), []byte("key")}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		db.Exec(cmdLine)
	}
}

// BenchmarkMGET 性能测试 MGET 命令
func BenchmarkMGET(b *testing.B) {
	db := MakeDB()
	// Prepare 100 keys
	for i := 0; i < 100; i++ {
		key := []byte(string(rune('a' + i)))
		db.Exec([][]byte{[]byte("SET"), key, []byte("value")})
	}

	cmdLine := [][]byte{[]byte("MGET")}
	for i := 0; i < 10; i++ {
		cmdLine = append(cmdLine, []byte(string(rune('a' + i))))
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		db.Exec(cmdLine)
	}
}

// BenchmarkConcurrentSET 并发 SET 性能测试
func BenchmarkConcurrentSET(b *testing.B) {
	db := MakeDB()
	cmdLine := [][]byte{[]byte("SET"), []byte("key"), []byte("value")}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			db.Exec(cmdLine)
		}
	})
}

// BenchmarkConcurrentGET 并发 GET 性能测试
func BenchmarkConcurrentGET(b *testing.B) {
	db := MakeDB()
	db.Exec([][]byte{[]byte("SET"), []byte("key"), []byte("value")})

	cmdLine := [][]byte{[]byte("GET"), []byte("key")}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			db.Exec(cmdLine)
		}
	})
}

// BenchmarkTTL SET with TTL 性能测试
func BenchmarkSETWithTTL(b *testing.B) {
	db := MakeDB()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// SET then EXPIRE
		db.Exec([][]byte{[]byte("SET"), []byte("key"), []byte("value")})
		db.Exec([][]byte{[]byte("EXPIRE"), []byte("key"), []byte("100")})
	}
}

// BenchmarkExpire 检查过期性能
func BenchmarkExpireCheck(b *testing.B) {
	db := MakeDB()
	db.Exec([][]byte{[]byte("SET"), []byte("key"), []byte("value")})
	db.Exec([][]byte{[]byte("EXPIRE"), []byte("key"), []byte("100")})

	cmdLine := [][]byte{[]byte("GET"), []byte("key")}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		db.Exec(cmdLine)
	}
}

// BenchmarkMemoryUsage 测试内存使用
func BenchmarkMemoryUsage(b *testing.B) {
	db := MakeDB()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		key := []byte(string(rune('a' + i%26)))
		value := []byte("test_value_with_some_data")
		db.Exec([][]byte{[]byte("SET"), key, value})
	}
}

// BenchmarkHashOperations Hash 操作性能测试
func BenchmarkHashHSET(b *testing.B) {
	db := MakeDB()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		db.Exec([][]byte{[]byte("HSET"), []byte("hash"), []byte("field"), []byte("value")})
	}
}

// BenchmarkHashHGET Hash GET 性能测试
func BenchmarkHashHGET(b *testing.B) {
	db := MakeDB()
	db.Exec([][]byte{[]byte("HSET"), []byte("hash"), []byte("field"), []byte("value")})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		db.Exec([][]byte{[]byte("HGET"), []byte("hash"), []byte("field")})
	}
}

// BenchmarkListOperations List 操作性能测试
func BenchmarkListLPUSH(b *testing.B) {
	db := MakeDB()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		db.Exec([][]byte{[]byte("LPUSH"), []byte("list"), []byte("value")})
	}
}

// BenchmarkSetOperations Set 操作性能测试
func BenchmarkSetSADD(b *testing.B) {
	db := MakeDB()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		db.Exec([][]byte{[]byte("SADD"), []byte("set"), []byte("member")})
	}
}

// BenchmarkSortedSetOperations Sorted Set 操作性能测试
func BenchmarkSortedSetZADD(b *testing.B) {
	db := MakeDB()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		db.Exec([][]byte{[]byte("ZADD"), []byte("zset"), []byte("1"), []byte("member")})
	}
}

// BenchmarkSlowLog 慢日志性能影响测试
func BenchmarkSlowLogImpact(b *testing.B) {
	db := MakeDB()
	cmdLine := [][]byte{[]byte("SET"), []byte("key"), []byte("value")}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		startTime := time.Now()
		db.Exec(cmdLine)
		duration := time.Since(startTime)
		db.AddSlowLogEntry(duration, cmdLine)
	}
}

// BenchmarkCommandParsing 命令解析性能测试
func BenchmarkCommandParsing(b *testing.B) {
	cmdLine := [][]byte{[]byte("SET"), []byte("mykey"), []byte("myvalue")}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		cmd := string(cmdLine[0])
		_ = cmd
	}
}
