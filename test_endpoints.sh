#!/bin/bash

# Test script to demonstrate the new HTTP endpoint functionality

echo "=== Modelplex HTTP Endpoint Testing ==="
echo

# Build modelplex
echo "Building modelplex..."
go build -o modelplex ./cmd/modelplex
echo "✓ Build complete"
echo

# Start server in background
echo "Starting modelplex server on port 11435..."
./modelplex --port 11435 &
SERVER_PID=$!
echo "✓ Server started (PID: $SERVER_PID)"
echo

# Wait for server to start
sleep 2

echo "Testing endpoints..."
echo

# Test health endpoint
echo "1. Health check:"
curl -s http://localhost:11435/health | jq '.'
echo

# Test models endpoint (new structure)
echo "2. Models endpoint (new structure /models/v1/models):"
curl -s http://localhost:11435/models/v1/models | jq '.data | length' | xargs echo "Available models:"
echo

# Test MCP tools endpoint
echo "3. MCP tools endpoint:"
curl -s http://localhost:11435/mcp/v1/tools | jq '.'
echo

# Test internal status endpoint
echo "4. Internal status endpoint:"
curl -s http://localhost:11435/_internal/status | jq '.'
echo

# Test backward compatibility
echo "5. Backward compatibility (old /v1/models):"
curl -s http://localhost:11435/v1/models | jq '.data | length' | xargs echo "Available models:"
echo

echo "✓ All endpoint tests completed"
echo

# Stop server
echo "Stopping server..."
kill $SERVER_PID
wait $SERVER_PID 2>/dev/null
echo "✓ Server stopped"
echo

echo "=== Testing socket mode ==="
echo

# Test socket mode
echo "Starting modelplex in socket mode..."
./modelplex --socket /tmp/test-modelplex.socket &
SOCKET_PID=$!
echo "✓ Socket server started (PID: $SOCKET_PID)"
echo

# Wait for server to start
sleep 2

echo "Testing socket endpoints..."

# Test health via socket
echo "1. Health check via socket:"
curl -s --unix-socket /tmp/test-modelplex.socket http://localhost/health | jq '.'
echo

# Test that internal endpoints are NOT available via socket
echo "2. Internal endpoints should not be available via socket:"
HTTP_CODE=$(curl -s -w "%{http_code}" --unix-socket /tmp/test-modelplex.socket http://localhost/_internal/status -o /dev/null)
if [ "$HTTP_CODE" = "404" ]; then
    echo "✓ Internal endpoints correctly blocked on socket (404)"
else
    echo "✗ Internal endpoints unexpectedly available on socket ($HTTP_CODE)"
fi
echo

# Stop socket server
echo "Stopping socket server..."
kill $SOCKET_PID
wait $SOCKET_PID 2>/dev/null
echo "✓ Socket server stopped"
echo

echo "=== All tests completed successfully! ==="
