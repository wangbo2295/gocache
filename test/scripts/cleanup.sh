#!/bin/bash

# GoCache Test Cleanup Script
# This script cleans up test data, logs, and temporary files

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Cleaning up test environment...${NC}"

# Stop server if running
if [ -f "./test/scripts/stop_server.sh" ]; then
    ./test/scripts/stop_server.sh
fi

# Remove test data directory
if [ -d "./test/data" ]; then
    echo "Removing test data directory..."
    rm -rf ./test/data
fi

# Remove log files
if [ -f "./test/gocache.log" ]; then
    echo "Removing log file..."
    rm -f ./test/gocache.log
fi

# Remove PID files
if [ -f "./test/gocache.pid" ]; then
    echo "Removing PID file..."
    rm -f ./test/gocache.pid
fi

# Remove persistence files
echo "Removing persistence files..."
rm -f ./test/*.aof
rm -f ./test/*.rdb
rm -f ./dump.rdb
rm -f ./appendonly.aof

# Remove test reports (optional)
if [ "$1" == "--reports" ]; then
    echo "Removing test reports..."
    rm -f ./test/reports/*.md
fi

# Remove binary
if [ -f "./test/bin/gocache" ]; then
    echo "Removing test binary..."
    rm -f ./test/bin/gocache
fi

echo -e "${GREEN}Cleanup complete${NC}"
echo ""
echo "Next steps:"
echo "  1. Start server: ./test/scripts/start_server.sh"
echo "  2. Run tests: go test ./test/e2e/... -v"
echo "  3. Stop server: ./test/scripts/stop_server.sh"
