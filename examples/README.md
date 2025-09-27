# gocapnweb Examples

This directory contains complete examples demonstrating the Go implementation of Cap'n Web RPC.

## Examples

### 1. Hello World (`helloworld/`)

A simple WebSocket RPC example that demonstrates:
- Basic method registration and dispatch
- WebSocket connection handling
- Real-time bidirectional communication
- Interactive web client

**Running:**
```bash
cd helloworld
go run main.go
```

**Demo:** http://localhost:8000/static/helloworld/

### 2. Batch Pipelining (`batch-pipelining/`)

An advanced HTTP batch RPC example that demonstrates:
- Pipeline references (chaining dependent calls)
- HTTP batch processing for optimal performance
- Simulated network latency comparison
- Performance metrics and visualization

**Running:**
```bash
cd batch-pipelining
go run main.go
```

**Demo:** http://localhost:8000/static/batch-pipelining/

## JavaScript Client

Both examples use the official `capnweb` JavaScript client library loaded from CDN:
```javascript
import { newWebSocketRpcSession, newHttpBatchRpcSession } from 'https://unpkg.com/capnweb@latest/dist/index.js';
```

For production use, install the package:
```bash
npm install capnweb
```

## File Structure

```
examples/
├── README.md                    # This file
├── static/                      # Static web assets
│   ├── helloworld/
│   │   ├── index.html          # Hello World demo page
│   │   └── client.mjs          # WebSocket client
│   └── batch-pipelining/
│       ├── index.html          # Batch pipelining demo page
│       └── client.mjs          # HTTP batch client
├── helloworld/
│   ├── go.mod                  # Go module
│   └── main.go                 # Hello World server
└── batch-pipelining/
    ├── go.mod                  # Go module
    └── main.go                 # Batch pipelining server
```

## Protocol Compatibility

These examples are fully compatible with the original C++ implementation:
- Same JSON message format
- Same WebSocket and HTTP endpoints
- Same pipeline reference syntax
- Same client-side JavaScript

## Testing the Examples

### Hello World Test
```bash
# Terminal 1: Start server
cd helloworld && go run main.go

# Terminal 2: Test with curl
curl -X POST http://localhost:8000/api \
  -d '["push",["pipeline",1,["hello"],["World"]]]'

curl -X POST http://localhost:8000/api \
  -d '["pull",1]'
```

### Batch Pipelining Test
```bash
# Terminal 1: Start server
cd batch-pipelining && go run main.go

# Terminal 2: Test authentication
curl -X POST http://localhost:8000/rpc \
  -d '["push",["pipeline",1,["authenticate"],["cookie-123"]]]'

curl -X POST http://localhost:8000/rpc \
  -d '["pull",1]'
```

## Performance Notes

The Go implementation provides:
- **Concurrent handling**: Each connection runs in its own goroutine
- **Memory safety**: No memory leaks or segfaults
- **Built-in JSON**: Fast, safe JSON processing
- **Standard HTTP**: Uses Go's battle-tested net/http package

Performance characteristics:
- **Latency**: Similar to C++ for most use cases
- **Throughput**: Excellent for typical RPC workloads
- **Memory**: Automatic garbage collection
- **Scalability**: Goroutines scale better than event loops for blocking operations

## Development

To modify the examples:

1. **Server changes**: Edit the `main.go` files
2. **Client changes**: Edit the `.mjs` files in `static/`
3. **UI changes**: Edit the `.html` files in `static/`

The servers automatically serve static files, so you can edit the client code and refresh the browser to see changes.

## Troubleshooting

### CORS Issues
The servers include CORS headers for browser compatibility. If you encounter CORS issues, check that:
- The server is running on the expected port
- The client is connecting to the correct URL
- No proxy is interfering with headers

### WebSocket Connection Issues
For WebSocket problems:
- Check browser developer tools for connection errors
- Verify the WebSocket URL (ws:// not https://)
- Ensure no firewall is blocking the connection

### Module Issues
If you see Go module errors:
```bash
go mod tidy
go mod download
```

### JavaScript Import Issues
The examples use ES modules with CDN imports. For older browsers or different setups, you may need to:
- Use a bundler like webpack or vite
- Install capnweb locally with npm
- Use a different import strategy
