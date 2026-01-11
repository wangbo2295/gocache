#!/bin/bash

# GoCache Test Runner Script
# This script runs all E2E tests and generates reports

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test categories
RUN_FUNCTIONAL=true
RUN_RELIABILITY=false
RUN_PERFORMANCE=true
RUN_STABILITY=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --functional)
            RUN_FUNCTIONAL=true
            shift
            ;;
        --no-functional)
            RUN_FUNCTIONAL=false
            shift
            ;;
        --reliability)
            RUN_RELIABILITY=true
            shift
            ;;
        --no-reliability)
            RUN_RELIABILITY=false
            shift
            ;;
        --performance)
            RUN_PERFORMANCE=true
            shift
            ;;
        --no-performance)
            RUN_PERFORMANCE=false
            shift
            ;;
        --stability)
            RUN_STABILITY=true
            shift
            ;;
        --no-stability)
            RUN_STABILITY=false
            shift
            ;;
        --all)
            RUN_FUNCTIONAL=true
            RUN_RELIABILITY=true
            RUN_PERFORMANCE=true
            RUN_STABILITY=true
            shift
            ;;
        --quick)
            RUN_FUNCTIONAL=true
            RUN_RELIABILITY=false
            RUN_PERFORMANCE=false
            RUN_STABILITY=false
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  --functional       Run functional tests (default: true)"
            echo "  --no-functional    Skip functional tests"
            echo "  --reliability      Run reliability tests (default: false)"
            echo "  --no-reliability   Skip reliability tests"
            echo "  --performance      Run performance tests (default: true)"
            echo "  --no-performance   Skip performance tests"
            echo "  --stability        Run stability tests (default: false)"
            echo "  --no-stability     Skip stability tests"
            echo "  --all              Run all tests"
            echo "  --quick            Run quick tests (functional only)"
            echo "  -h, --help         Show this help message"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            exit 1
            ;;
    esac
done

# Create reports directory
mkdir -p ./test/reports

# Start server
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}GoCache E2E Test Suite${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

./test/scripts/start_server.sh
sleep 2

# Trap to ensure server is stopped on exit
trap "./test/scripts/stop_server.sh" EXIT

# Test counter
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to run tests
run_tests() {
    local category=$1
    local path=$2
    local description=$3

    echo -e "${BLUE}----------------------------------------${NC}"
    echo -e "${BLUE}Running: $description${NC}"
    echo -e "${BLUE}Category: $category${NC}"
    echo -e "${BLUE}----------------------------------------${NC}"

    local log_file="./test/reports/${category}.log"
    local start_time=$(date +%s)

    if go test "$path" -v -timeout 30m 2>&1 | tee "$log_file"; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        echo -e "${GREEN}✓ $description passed (${duration}s)${NC}"
        ((PASSED_TESTS++))
    else
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        echo -e "${RED}✗ $description failed (${duration}s)${NC}"
        echo -e "${RED}See $log_file for details${NC}"
        ((FAILED_TESTS++))
    fi
    ((TOTAL_TESTS++))
    echo ""
}

# Run selected test suites
if [ "$RUN_FUNCTIONAL" = true ]; then
    run_tests "functional" "./test/e2e/functional/..." "Functional Tests"

    # Individual functional tests
    run_tests "string" "./test/e2e/functional -run TestString" "String Commands"
    run_tests "hash" "./test/e2e/functional -run TestHash" "Hash Commands"
    run_tests "list" "./test/e2e/functional -run TestList" "List Commands"
    run_tests "set" "./test/e2e/functional -run TestSet" "Set Commands"
    run_tests "sortedset" "./test/e2e/functional -run TestSortedSet" "SortedSet Commands"
    run_tests "ttl" "./test/e2e/functional -run TestTTL" "TTL Commands"
    run_tests "transaction" "./test/e2e/functional -run TestTransaction" "Transaction Commands"
fi

if [ "$RUN_RELIABILITY" = true ]; then
    run_tests "reliability" "./test/e2e/reliability/..." "Reliability Tests"
fi

if [ "$RUN_PERFORMANCE" = true ]; then
    run_tests "performance" "./test/e2e/performance/..." "Performance Tests"

    # Run benchmarks
    echo -e "${BLUE}----------------------------------------${NC}"
    echo -e "${BLUE}Running: Performance Benchmarks${NC}"
    echo -e "${BLUE}----------------------------------------${NC}"

    local bench_log="./test/reports/benchmark.log"
    if go test ./test/e2e/performance/... -bench=. -benchmem -timeout 30m 2>&1 | tee "$bench_log"; then
        echo -e "${GREEN}✓ Benchmarks completed${NC}"
    else
        echo -e "${YELLOW}⚠ Some benchmarks may have failed${NC}"
    fi
    echo ""
fi

if [ "$RUN_STABILITY" = true ]; then
    run_tests "stability" "./test/e2e/stability/..." "Stability Tests"
fi

# Print summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "Total test suites: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
if [ $FAILED_TESTS -gt 0 ]; then
    echo -e "${RED}Failed: $FAILED_TESTS${NC}"
fi
echo ""

# Generate reports
if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    echo "See ./test/reports/ for detailed logs"
else
    echo -e "${RED}Some tests failed${NC}"
    echo "Check ./test/reports/*.log for details"
    exit 1
fi
