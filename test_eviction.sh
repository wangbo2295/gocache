#!/bin/bash

# Test script for memory eviction functionality

echo "========================================="
echo "GoCache Memory Eviction Test"
echo "========================================="
echo ""

# Start the server in background
echo "Starting GoCache server with 100kb memory limit and LRU eviction..."
./goredis -c test_eviction.conf &
SERVER_PID=$!
echo "Server PID: $SERVER_PID"
sleep 2

# Test commands
echo ""
echo "1. Testing INFO command..."
redis-cli -p 16379 INFO | grep -A 5 "Memory"

echo ""
echo "2. Testing MEMORY STATS command..."
redis-cli -p 16379 MEMORY STATS

echo ""
echo "3. Setting keys (should trigger eviction after 100kb)..."
for i in {1..20}; do
  VALUE=$(openssl rand -base64 10000) # ~13KB per value
  redis-cli -p 16379 SET "key$i" "$VALUE" > /dev/null
  echo "   Set key$i"
done

echo ""
echo "4. Checking which keys survived (LRU should have evicted some)..."
echo "   Checking keys 1-10:"
for i in {1..10}; do
  EXISTS=$(redis-cli -p 16379 EXISTS "key$i")
  if [ "$EXISTS" == "1" ]; then
    echo "   key$i: EXISTS"
  else
    echo "   key$i: EVICTED"
  fi
done

echo ""
echo "   Checking keys 11-20 (should all exist - recently added):"
for i in {11..20}; do
  EXISTS=$(redis-cli -p 16379 EXISTS "key$i")
  if [ "$EXISTS" == "1" ]; then
    echo "   key$i: EXISTS"
  else
    echo "   key$i: EVICTED"
  fi
done

echo ""
echo "5. Checking memory usage..."
redis-cli -p 16379 MEMORY STATS

echo ""
echo "6. Testing MEMORY USAGE command for individual keys..."
redis-cli -p 16379 MEMORY USAGE key11
redis-cli -p 16379 MEMORY USAGE key20

echo ""
echo "7. Accessing key11 (making it more recently used)..."
redis-cli -p 16379 GET key11 > /dev/null
echo "   Accessed key11"

echo ""
echo "8. Adding more keys to trigger eviction of old keys..."
for i in {21..25}; do
  VALUE=$(openssl rand -base64 10000)
  redis-cli -p 16379 SET "key$i" "$VALUE" > /dev/null
  echo "   Set key$i"
done

echo ""
echo "9. Checking if key11 survived (it was accessed recently)..."
EXISTS=$(redis-cli -p 16379 EXISTS "key11")
if [ "$EXISTS" == "1" ]; then
  echo "   key11: EXISTS (good - LRU working!)"
else
  echo "   key11: EVICTED (unexpected)"
fi

echo ""
echo "10. Final memory stats..."
redis-cli -p 16379 INFO | grep -A 5 "Memory"

echo ""
echo "========================================="
echo "Test complete!"
echo "========================================="
echo ""

# Shutdown server
echo "Shutting down server..."
kill $SERVER_PID
wait $SERVER_PID 2>/dev/null

# Cleanup
echo "Cleaning up test files..."
rm -f test_eviction.aof

echo "Done!"
