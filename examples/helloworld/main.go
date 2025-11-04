package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/gocapnweb"
)

// HelloServer implements a simple "hello" RPC method.
type HelloServer struct {
	*gocapnweb.BaseRpcTarget
}

// NewHelloServer creates a new HelloServer instance.
func NewHelloServer() *HelloServer {
	server := &HelloServer{
		BaseRpcTarget: gocapnweb.NewBaseRpcTarget(),
	}

	// Register the hello method
	server.Method("hello", func(args json.RawMessage) (interface{}, error) {
		// Parse arguments as array of strings
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
	// Default to serving static files from the examples/static directory
	staticPath := "/static"
	if len(os.Args) >= 2 {
		staticPath = os.Args[1]
	}

	port := ":8000"

	// Create Echo server with middleware
	e := gocapnweb.SetupEchoServer()

	// Setup RPC endpoint
	server := NewHelloServer()
	gocapnweb.SetupRpcEndpoint(e, "/api", server)

	// Setup static file endpoint
	gocapnweb.SetupFileEndpoint(e, "/static", staticPath)

	log.Printf("🚀 Hello World Go Server (Echo) starting on port %s", port)
	log.Printf("📁 Static files served from: %s", staticPath)
	log.Printf("🔌 WebSocket RPC endpoint: ws://localhost%s/api", port)
	log.Printf("📄 Static files: http://localhost%s/static/", port)
	log.Printf("🌐 Demo URL: http://localhost%s/static/helloworld/", port)
	log.Println()
	log.Println("Try the demo:")
	log.Printf("  curl -X POST http://localhost%s/api -d '[\"push\",[\"pipeline\",1,[\"hello\"],[\"World\"]]]'", port)
	log.Printf("  curl -X POST http://localhost%s/api -d '[\"pull\",1]'", port)

	if err := e.Start(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
