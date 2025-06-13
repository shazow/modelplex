#!/bin/bash

# Test script for Modelplex API using curl
# Requires socat to communicate with Unix socket over HTTP

SOCKET_PATH="./modelplex.socket"

echo "🧪 Testing Modelplex API"

# Check if socket exists
if [ ! -S "$SOCKET_PATH" ]; then
    echo "❌ Socket not found at $SOCKET_PATH"
    echo "   Make sure Modelplex is running"
    exit 1
fi

# Test health endpoint
echo "📡 Testing health endpoint..."
socat - "UNIX-CONNECT:$SOCKET_PATH" << 'EOF'
GET /health HTTP/1.1
Host: localhost
Connection: close

EOF

echo -e "\n"

# Test models endpoint
echo "📋 Testing models endpoint..."
socat - "UNIX-CONNECT:$SOCKET_PATH" << 'EOF'
GET /v1/models HTTP/1.1
Host: localhost
Content-Type: application/json
Connection: close

EOF

echo -e "\n"

# Test chat completions
echo "💬 Testing chat completions..."
socat - "UNIX-CONNECT:$SOCKET_PATH" << 'EOF'
POST /v1/chat/completions HTTP/1.1
Host: localhost
Content-Type: application/json
Connection: close
Content-Length: 200

{
  "model": "gpt-4",
  "messages": [
    {"role": "user", "content": "Hello from isolated environment!"}
  ],
  "max_tokens": 50
}
EOF

echo -e "\n✅ API tests completed"