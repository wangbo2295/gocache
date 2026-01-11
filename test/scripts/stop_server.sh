#!/bin/bash

# GoCache Test Server Stop Script
# This script stops the GoCache server

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PIDFILE="./test/gocache.pid"

# Check if PID file exists
if [ ! -f "$PIDFILE" ]; then
    echo -e "${YELLOW}No PID file found. Checking for running processes...${NC}"

    # Try to find GoCache process
    PIDS=$(pgrep -f "gocache" || true)

    if [ -z "$PIDS" ]; then
        echo -e "${YELLOW}No GoCache server running${NC}"
        exit 0
    fi

    echo -e "${YELLOW}Found GoCache processes: $PIDS${NC}"
    read -p "Kill these processes? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "$PIDS" | xargs kill
        echo -e "${GREEN}Server stopped${NC}"
    fi
    exit 0
fi

# Read PID from file
PID=$(cat "$PIDFILE")

# Check if process is running
if ! ps -p "$PID" > /dev/null 2>&1; then
    echo -e "${YELLOW}Server is not running (stale PID file)${NC}"
    rm -f "$PIDFILE"
    exit 0
fi

echo -e "${YELLOW}Stopping GoCache server (PID: $PID)...${NC}"

# Try graceful shutdown first
kill "$PID" 2>/dev/null || true

# Wait for process to stop
for i in {1..10}; do
    if ! ps -p "$PID" > /dev/null 2>&1; then
        echo -e "${GREEN}Server stopped successfully${NC}"
        rm -f "$PIDFILE"
        exit 0
    fi
    sleep 1
done

# If still running, force kill
echo -e "${YELLOW}Server did not stop gracefully, forcing...${NC}"
kill -9 "$PID" 2>/dev/null || true

# Final check
if ps -p "$PID" > /dev/null 2>&1; then
    echo -e "${RED}Failed to stop server${NC}"
    exit 1
else
    echo -e "${GREEN}Server stopped (forced)${NC}"
    rm -f "$PIDFILE"
fi
