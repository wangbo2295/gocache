#!/bin/bash

# Test script for GoCache server

# Start server in background
./bin/gocache > /tmp/server.log 2>&1 &
SERVER_PID=$!

# Wait for server to start
sleep 2

# Check if server is running
if ! kill -0 $SERVER_PID 2>/dev/null; then
    echo "Failed to start server"
    cat /tmp/server.log
    exit 1
fi

echo "✓ Server started (PID: $SERVER_PID)"

# Function to send command and read response
send_cmd() {
    local cmd="$1"
    {
        echo "$cmd"
        sleep 0.1
    } | telnet 127.0.0.1 6379 2>&1 | grep -A 1 "Escape character" | tail -1 | sed 's/\r$//'
}

# Test PING
echo -n "Testing PING... "
response=$(send_cmd "PING")
if [ "$response" = "+PONG" ]; then
    echo "✓ PASS"
else
    echo "✗ FAIL (got: $response)"
fi

# Test SET
echo -n "Testing SET... "
response=$(send_cmd "SET mykey myvalue")
if [ "$response" = "+OK" ]; then
    echo "✓ PASS"
else
    echo "✗ FAIL (got: $response)"
fi

# Test GET
echo -n "Testing GET... "
response=$(send_cmd "GET mykey")
if [[ "$response" == *"myvalue"* ]]; then
    echo "✓ PASS"
else
    echo "✗ FAIL (got: $response)"
fi

# Test INCR
echo -n "Testing INCR... "
response=$(send_cmd "SET counter 0" > /dev/null; send_cmd "INCR counter")
if [ "$response" = ":1" ]; then
    echo "✓ PASS"
else
    echo "✗ FAIL (got: $response)"
fi

# Cleanup
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo ""
echo "All tests completed!"
