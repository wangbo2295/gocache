# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GoCache is a Redis-compatible in-memory cache system implemented in Go. It's a learning project designed to understand distributed systems concepts while maintaining compatibility with the Redis RESP protocol. The project follows a clean, layered architecture with clear separation of concerns.

**Current Status**: MVP (v1.0.0) - Core functionality implemented including TCP server, RESP protocol, all major data structures (String, Hash, List, Set, SortedSet), AOF persistence, and TTL support.

## Build and Test Commands

### Build
```bash
go build -o bin/gocache
# Or simply:
go build
```

### Run
```bash
# With default config
./gocache

# With custom config file
./gocache -c gocache.conf
```

### Test
```bash
# All tests
go test ./... -v

# With coverage
go test ./... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out

# Race detection
go test ./... -race

# Single package test
go test ./database -v
```

### Code Quality
```bash
# Format code
gofmt -l .

# Static analysis
go vet ./...
```

## Architecture Overview

The system follows a layered architecture with clear boundaries:

```
Client → TCP Server → Handler → DB → ConcurrentDict
               ↓         ↓
         RESP Parser   AOF
```

### Request Flow

1. **Client Connection**: TCP server accepts connections (`server/server.go`)
2. **Protocol Parsing**: RESP parser converts bytes to command arrays (`protocol/resp/parser.go`)
3. **Command Routing**: Handler routes commands to database operations (`server/server.go` - `ExecCommand`)
4. **Execution**: Database executes commands using concurrent dictionary (`database/db.go`)
5. **Persistence**: Write operations are logged to AOF if enabled (`persistence/aof/`)
6. **Response**: Results converted to RESP format and sent to client

### Key Components

**Server Layer** (`server/`)
- TCP server with connection pooling
- Handler pattern for command execution
- Integrates AOF persistence transparently

**Protocol Layer** (`protocol/`)
- `protocol/commands.go`: Command definitions and classification (write/integer/status commands)
- `protocol/resp/`: Full RESP protocol implementation (parser and reply builders)
- Case-insensitive command handling via `ToUpper()`

**Database Engine** (`database/`)
- `db.go`: Core database with data, TTL, and version dictionaries
- Time wheel for efficient TTL management (10ms intervals, 1024 buckets)
- Transaction support (MULTI/EXEC) with optimistic locking (WATCH)
- Memory tracking and eviction policy integration
- Lazy TTL deletion (checked on access, not background scan)

**Storage Engine** (`dict/`)
- `ConcurrentDict`: Sharded concurrent dictionary with 16 shards (power of 2)
- FNV hash-based sharding with bit manipulation for shard selection
- Fine-grained locking reduces contention for parallel access

**Data Structures** (`datastruct/`)
- String, Hash (map-based), List (linked list), Set (map-based), SortedSet (skiplist + map)
- `TimeWheel`: Hierarchical time wheel for efficient TTL expiration

**Persistence** (`persistence/`)
- `aof/`: Append-only file persistence with configurable fsync strategies
- `rdb/`: RDB snapshot framework (not fully implemented)

**Eviction** (`eviction/`, `evictionpkg/`)
- LRU and LFU implementations
- Configurable policies: allkeys-lru/lfu, volatile-lru/lfu, etc.

## Important Architecture Patterns

### Sharded Locking
The `ConcurrentDict` uses 16 shards (power of 2 aligned) with separate RWMutex locks. The shard index is calculated using FNV hash with bit manipulation. This allows concurrent reads/writes to different keys without lock contention.

### Time Wheel TTL
Instead of background scanning, TTL is managed through a hierarchical time wheel (10ms tick, 1024 buckets). Keys are added to the wheel with callbacks for expiration. This is more efficient than periodic scanning.

### Handler Pattern
The `Handler` struct embeds both DB and AOF handler. When executing commands:
1. Command is executed in DB
2. If AOF enabled AND command is a write operation, log to AOF
3. AOF errors don't fail commands (graceful degradation)

### RESP Protocol
The system implements full Redis RESP protocol with support for:
- Simple Strings (`+OK\r\n`)
- Errors (`-ERR ...\r\n`)
- Integers (`:123\r\n`)
- Bulk Strings (`$5\r\nhello\r\n`)
- Arrays (`*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n`)

Case-insensitive commands are achieved through ASCII-only `ToUpper()` for performance.

### Memory Management
- Atomic counters track memory usage per database
- Memory limit checks trigger eviction before insertions
- Eviction policies are pluggable (LRU/LFU currently implemented)

## Configuration

Configuration is loaded from `.conf` files (Redis-compatible format):

```conf
bind 127.0.0.1
port 6379
databases 16
appendonly yes
appendfsync everysec
maxmemory 256mb
maxmemory-policy allkeys-lru
```

The `config` package provides a global `Config` instance that is initialized at startup. See `config/` for all available options.

## Testing Strategy

The project maintains high test coverage (86.2% average). Tests are organized by package:
- Unit tests for core components (dict, datastruct, database)
- Integration tests for protocol parsing
- End-to-end tests in `test/` directory

When adding new features:
1. Write unit tests in the same package
2. Test concurrency with `-race` flag
3. Ensure coverage ≥80% for new code

## Common Patterns

### Adding a New Command

1. Add command constant to `protocol/commands.go`
2. Classify as write/integer/status command in respective maps
3. Implement command logic in `database/db.go` (or appropriate data structure file)
4. Update `WriteCommands`/`IntegerCommands`/`StatusCommands` maps if needed

### Working with ConcurrentDict

```go
// Get
val, ok := dict.Get(key)

// Put (returns 1 if new key, 0 if update)
result := dict.Put(key, value)

// Remove
val, result := dict.Remove(key)

// ForEach with callback
dict.ForEach(func(key string, val interface{}) bool {
    // return false to stop iteration
    return true
})
```

### AOF Integration

Write operations are automatically logged to AOF if enabled. The handler checks `protocol.IsWriteCommand()` to determine if logging is needed. AOF errors are logged but don't fail commands.

## Module Dependencies

```
main.go
  ├── config (configuration loading)
  ├── logger (logging setup)
  ├── database (DB creation)
  │   ├── dict (concurrent dictionary)
  │   ├── datastruct (data structures)
  │   └── eviction (eviction policies)
  ├── persistence/aof (AOF handler)
  └── server (TCP server + handler)
      ├── protocol/resp (parsing)
      └── database (command execution)
```

## Performance Targets

- **QPS**: ≥50,000 (MVP), ≥100,000 (final)
- **P99 Latency**: <5ms (MVP), <1ms (final)
- **Concurrent Connections**: 10,000+

## Roadmap

Current version (v1.0.0-MVP) implements core features. Planned iterations:

1. ✅ Data structures (Hash, List, Set, SortedSet)
2. ✅ Eviction policies (LRU, LFU)
3. 🔄 AOF rewrite
4. 📋 RDB snapshot
5. 📋 Replication
6. 📋 Cluster mode

## Important Notes

- **No Redis/Godis terminology in docs**: When creating design documentation, use generic "内存缓存系统" terminology
- **Go version**: Uses Go 1.23+ features
- **Testing**: Always run with `-race` when testing concurrent code
- **Memory tracking**: Use `db.addMemoryUsage()` when inserting data to maintain accurate memory accounting
- **TTL handling**: Use lazy deletion (check on access) combined with time wheel for active expiration
