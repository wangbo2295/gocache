# GoCache E2E Testing Guide

This guide provides step-by-step instructions for running end-to-end acceptance tests for the GoCache project.

## Prerequisites

### System Requirements
- **OS**: macOS, Linux, or Windows with WSL
- **Go**: 1.23 or higher
- **Memory**: 8GB+ recommended
- **CPU**: 4+ cores recommended

### Software Requirements
- Go compiler and toolchain
- Git (for version control)
- redis-cli (optional, for manual testing)

## Quick Start

### 1. Build the Server

```bash
cd /Users/wangbo/goredis
go build -o gocache .
```

### 2. Start the Test Server

```bash
# Using the helper script
./test/scripts/start_server.sh

# Or manually
./gocache
```

The server will:
- Listen on `127.0.0.1:6379`
- Create data directory at `./test/data/`
- Log to `./test/gocache.log`

### 3. Run Tests

```bash
# Run all tests (functional + performance)
./test/scripts/run_all_tests.sh

# Run only functional tests
./test/scripts/run_all_tests.sh --quick

# Run with custom options
./test/scripts/run_all_tests.sh --performance --no-functional
```

### 4. Stop the Server

```bash
./test/scripts/stop_server.sh
```

## Test Categories

### Functional Tests (`test/e2e/functional/`)

Tests all data types and commands:

```bash
# Run all functional tests
go test ./test/e2e/functional/... -v

# Run specific data type tests
go test ./test/e2e/functional -run TestString -v
go test ./test/e2e/functional -run TestHash -v
go test ./test/e2e/functional -run TestList -v
go test ./test/e2e/functional -run TestSet -v
go test ./test/e2e/functional -run TestSortedSet -v
go test ./test/e2e/functional -run TestTTL -v
go test ./test/e2e/functional -run TestTransaction -v
```

**Coverage**:
- String: 17 commands
- Hash: 11 commands
- List: 11 commands
- Set: 14 commands
- SortedSet: 9 commands
- TTL: 7 commands
- Transaction: 4 commands

### Performance Tests (`test/e2e/performance/`)

Tests QPS, latency, and concurrency:

```bash
# Run all performance tests
go test ./test/e2e/performance/... -v

# Run specific test categories
go test ./test/e2e/performance -run TestQPS -v
go test ./test/e2e/performance -run TestLatency -v
go test ./test/e2e/performance -run TestConcurrent -v

# Run benchmarks
go test ./test/e2e/performance/... -bench=. -benchmem
```

**Metrics**:
- QPS: ≥ 80,000 (acceptance)
- P99 Latency: < 2ms (acceptance)
- Concurrent Connections: ≥ 5,000 (acceptance)

## Test Scripts

### start_server.sh
Starts the GoCache server for testing.

```bash
./test/scripts/start_server.sh [OPTIONS]

Options:
  -c, --config FILE    Use specified config file
  -p, --port PORT      Port to listen on (default: 6379)
  -b, --bind ADDRESS   Address to bind to (default: 127.0.0.1)
```

### stop_server.sh
Stops the GoCache server.

```bash
./test/scripts/stop_server.sh
```

### cleanup.sh
Cleans up test data, logs, and temporary files.

```bash
./test/scripts/cleanup.sh [--reports]
```

### run_all_tests.sh
Runs all or selected test suites.

```bash
./test/scripts/run_all_tests.sh [OPTIONS]

Options:
  --functional       Run functional tests (default: true)
  --reliability      Run reliability tests (default: false)
  --performance      Run performance tests (default: true)
  --stability        Run stability tests (default: false)
  --all              Run all tests
  --quick            Run quick tests (functional only)
```

## Test Reports

After running tests, reports are generated in `test/reports/`:

- `functional_report.md`: Functional test results
- `performance_report.md`: Performance metrics
- `acceptance_report.md`: Final acceptance status

## Troubleshooting

### Server Won't Start

**Issue**: Port already in use

```bash
# Check what's using port 6379
lsof -i :6379

# Or use a different port
./test/scripts/start_server.sh -p 6380
```

**Issue**: Build errors

```bash
# Clean and rebuild
go clean -cache
go mod tidy
go build -o gocache .
```

### Tests Fail to Connect

**Issue**: Server not running

```bash
# Start server
./test/scripts/start_server.sh

# Verify it's running
ps aux | grep gocache

# Test connection
redis-cli ping  # Should return PONG
```

**Issue**: Wrong port

```bash
# Ensure server is on 6379
./test/scripts/start_server.sh -p 6379
```

### Performance Tests Show Low QPS

**Issue**: Background processes

```bash
# Close unnecessary applications
# Use dedicated test machine
# Run multiple times and average
```

**Issue**: Resource limits

```bash
# Check available memory
free -h  # Linux
vm_stat  # macOS

# Check CPU load
top -n 1
```

### Tests Timeout

**Issue**: Tests taking too long

```bash
# Increase timeout
go test ./test/e2e/... -timeout 60m

# Or run shorter tests
./test/scripts/run_all_tests.sh --quick
```

## Continuous Integration

### GitHub Actions Example

```yaml
name: E2E Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.23
      - name: Build
        run: go build -o gocache .
      - name: Start Server
        run: ./test/scripts/start_server.sh
      - name: Run Tests
        run: ./test/scripts/run_all_tests.sh --quick
      - name: Stop Server
        if: always()
        run: ./test/scripts/stop_server.sh
```

## Best Practices

1. **Before Testing**
   - Ensure clean environment: `./test/scripts/cleanup.sh`
   - Update dependencies: `go mod tidy`
   - Rebuild server: `go build`

2. **During Testing**
   - Run tests in order: functional → reliability → performance
   - Monitor system resources
   - Check logs if tests fail: `tail -f test/gocache.log`

3. **After Testing**
   - Review test reports
   - Check for resource leaks
   - Clean up test data

## Acceptance Criteria

### Functional (100% Required)
- ✅ All data types working correctly
- ✅ All commands execute properly
- ✅ TTL expiration works
- ✅ Transactions maintain atomicity
- ✅ Binary data handled correctly

### Performance (80% of Target)
- ✅ QPS ≥ 80,000
- ✅ P99 Latency < 2ms
- ✅ 5,000+ concurrent connections

### Reliability (100% Required)
- ✅ Data persistence works
- ✅ Recovery after crash
- ✅ No data corruption

## Support

For issues or questions:
1. Check logs: `test/gocache.log`
2. Review test reports: `test/reports/`
3. Check troubleshooting section above

---

**Happy Testing! 🚀**
