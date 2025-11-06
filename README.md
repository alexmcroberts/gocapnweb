# gocapnweb

[Cap'n Web](https://github.com/cloudflare/capnweb) Go Server Library. This library allows you to create server implementations for the Cap'n Web RPC protocol in Go, providing the minimal plumbing required for getting messages flying between RPC clients and servers.

## Features

- **WebSocket RPC**: Real-time bidirectional RPC over WebSockets
- **HTTP Batch RPC**: Batched RPC calls over HTTP POST for optimal pipelining performance
- **Pipeline References**: Chain RPC calls where one call's result is used as input to another
- **Static File Serving**: Serve static files with proper MIME types and security checks
- **Goroutine-Safe**: Thread-safe session management with proper synchronization

## Dependencies

- Go 1.21+
- [gorilla/websocket](https://github.com/gorilla/websocket) for WebSocket support
- [gorilla/mux](https://github.com/gorilla/mux) for HTTP routing

## Installation

```bash
go get github.com/gocapnweb
```

## Examples

### Running the Examples

1. **Hello World**:
   ```bash
   cd examples/helloworld
   go run main.go
   ```

   In a second tab:
   ```bash
   cd static
   npm install
   npm run dev
   ```
   Open: http://localhost:3000

2. **Batch Pipelining**:
   ```bash
   cd examples/batch-pipelining  
   go run main.go
   ```

   In a second tab:
   ```bash
   cd static
   npm install
   npm run dev
   ```
   Open: http://localhost:3000

3. **Server Push**
   ```bash
   cd examples/serverpush
   go run main.go
   ```

   In a second tab:
   ```bash
   cd static
   npm install
   npm run dev
   ```
   Open: http://localhost:3000

4. **Bluesky**
   ```bash
   cd examples/bluesky
   go run main.go
   ```

   In a second tab:
   ```bash
   cd static
   npm install
   npm run dev
   ```
   Open: http://localhost:3000

## Getting Started

### Simple Hello World Server

```go
package main

import (
    "encoding/json"
    "log"
    "net/http"
    
    "github.com/gocapnweb"
    "github.com/gorilla/mux"
)

// Create a simple RPC target
type HelloServer struct {
    *gocapnweb.BaseRpcTarget
}

func NewHelloServer() *HelloServer {
    server := &HelloServer{
        BaseRpcTarget: gocapnweb.NewBaseRpcTarget(),
    }
    
    // Register the hello method
    server.Method("hello", func(args json.RawMessage) (interface{}, error) {
        var argArray []string
        if err := json.Unmarshal(args, &argArray); err != nil {
            return nil, err
        }
        
        if len(argArray) == 0 {
            return "Hello, World!", nil
        }
        
        return "Hello, " + argArray[0] + "!", nil
    })
    
    return server
}

func main() {
    router := mux.NewRouter()
    
    // Setup RPC endpoint
    server := NewHelloServer()
    gocapnweb.SetupRpcEndpoint(router, "/api", server)
    
    // Setup static file serving
    gocapnweb.SetupFileEndpoint(router, "/static", "./static")
    
    log.Println("Server starting on :8000")
    log.Fatal(http.ListenAndServe(":8000", router))
}
```

### Using Custom RPC Target

You can implement the `RpcTarget` interface directly for more control:

```go
type CustomServer struct {
    // your fields
}

func (s *CustomServer) Dispatch(method string, args json.RawMessage) (interface{}, error) {
    switch method {
    case "myMethod":
        return s.handleMyMethod(args)
    default:
        return nil, fmt.Errorf("method not found: %s", method)
    }
}
```

## Key Components

### RpcTarget Interface

The core interface that your server must implement:

```go
type RpcTarget interface {
    Dispatch(method string, args json.RawMessage) (interface{}, error)
}
```

### BaseRpcTarget

A convenient base implementation with method registration:

```go
server := gocapnweb.NewBaseRpcTarget()
server.Method("methodName", handlerFunc)
```

### Session Management

Each WebSocket connection or HTTP batch request gets its own session with:
- Pipeline reference resolution
- Export ID management
- Result caching
- Thread-safe operations

## Protocol Support

### WebSocket RPC

Real-time bidirectional communication:
- Connect to `ws://yourserver/api`
- Send/receive JSON-RPC messages
- Automatic session management

### HTTP Batch RPC

Optimized for pipelining multiple dependent calls:
- POST to `/api` with newline-separated JSON messages
- Single round trip for multiple dependent operations
- Automatic pipeline reference resolution

### Pipeline References

Chain operations efficiently:

```javascript
// Client-side example
const api = newHttpBatchRpcSession('http://localhost:8000/rpc');
const user = api.authenticate('session-token');
const profile = api.getUserProfile(user.id);  // user.id resolved automatically
const notifications = api.getNotifications(user.id);

// All three calls made in a single HTTP request!
const [u, p, n] = await Promise.all([user, profile, notifications]);
```
