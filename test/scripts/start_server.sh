#!/bin/bash

# GoCache Test Server Startup Script
# This script starts the GoCache server for testing

set -e

# Default values
BIND="127.0.0.1"
PORT=6379
DATADIR="./test/data"
LOGFILE="./test/gocache.log"
PIDFILE="./test/gocache.pid"
CONFIG_FILE=""

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        -p|--port)
            PORT="$2"
            shift 2
            ;;
        -b|--bind)
            BIND="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  -c, --config FILE    Use specified config file"
            echo "  -p, --port PORT      Port to listen on (default: 6379)"
            echo "  -b, --bind ADDRESS   Address to bind to (default: 127.0.0.1)"
            echo "  -h, --help           Show this help message"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            exit 1
            ;;
    esac
done

# Create data directory if it doesn't exist
mkdir -p "$DATADIR"

# Check if server is already running
if [ -f "$PIDFILE" ]; then
    PID=$(cat "$PIDFILE")
    if ps -p "$PID" > /dev/null 2>&1; then
        echo -e "${YELLOW}Server is already running (PID: $PID)${NC}"
        exit 0
    else
        echo -e "${YELLOW}Removing stale PID file${NC}"
        rm -f "$PIDFILE"
    fi
fi

echo -e "${GREEN}Starting GoCache server...${NC}"

# Build the server if not already built
if [ ! -f "./gocache" ] && [ ! -f "./bin/gocache" ]; then
    echo -e "${YELLOW}Building GoCache server...${NC}"
    go build -o ./gocache . || {
        echo -e "${RED}Failed to build server${NC}"
        exit 1
    }
fi

# Determine server binary
if [ -f "./gocache" ]; then
    SERVER="./gocache"
elif [ -f "./bin/gocache" ]; then
    SERVER="./bin/gocache"
else
    echo -e "${RED}Server binary not found${NC}"
    exit 1
fi

# Start the server
if [ -n "$CONFIG_FILE" ]; then
    echo "Using config file: $CONFIG_FILE"
    nohup "$SERVER" -c "$CONFIG_FILE" > "$LOGFILE" 2>&1 &
else
    echo "Listening on $BIND:$PORT"
    nohup "$SERVER" > "$LOGFILE" 2>&1 &
    PID=$!
    echo $PID > "$PIDFILE"
fi

# Wait for server to start
sleep 1

# Check if server started successfully
if [ -f "$PIDFILE" ]; then
    PID=$(cat "$PIDFILE")
    if ps -p "$PID" > /dev/null 2>&1; then
        echo -e "${GREEN}Server started successfully (PID: $PID)${NC}"
        echo "Log file: $LOGFILE"
        echo "Data directory: $DATADIR"

        # Test connection
        sleep 1
        if command -v redis-cli &> /dev/null; then
            if redis-cli -p "$PORT" ping 2>/dev/null | grep -q "PONG"; then
                echo -e "${GREEN}Server is responding to commands${NC}"
            else
                echo -e "${YELLOW}Warning: Server may not be ready yet${NC}"
            fi
        fi
    else
        echo -e "${RED}Failed to start server${NC}"
        rm -f "$PIDFILE"
        exit 1
    fi
else
    # If using config file with built-in PID management
    if pgrep -f "$SERVER" > /dev/null; then
        echo -e "${GREEN}Server started successfully${NC}"
        echo "Log file: $LOGFILE"
    else
        echo -e "${RED}Failed to start server${NC}"
        exit 1
    fi
fi
