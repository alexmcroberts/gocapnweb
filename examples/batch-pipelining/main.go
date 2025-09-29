package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gocapnweb"
)

// User represents a user object.
type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Profile represents a user profile.
type Profile struct {
	ID  string `json:"id"`
	Bio string `json:"bio"`
}

// UserServer implements RPC methods for user operations with pipelining support.
type UserServer struct {
	*gocapnweb.BaseRpcTarget
	users         map[string]User     // session token -> user
	profiles      map[string]Profile  // user ID -> profile
	notifications map[string][]string // user ID -> notifications
}

// NewUserServer creates a new UserServer instance with sample data.
func NewUserServer() *UserServer {
	server := &UserServer{
		BaseRpcTarget: gocapnweb.NewBaseRpcTarget(),
		users:         make(map[string]User),
		profiles:      make(map[string]Profile),
		notifications: make(map[string][]string),
	}

	// Initialize sample data
	server.initializeData()

	// Register RPC methods
	server.Method("authenticate", server.authenticate)
	server.Method("getUserProfile", server.getUserProfile)
	server.Method("getNotifications", server.getNotifications)

	return server
}

func (s *UserServer) initializeData() {
	// Initialize users (session token -> user object)
	s.users["cookie-123"] = User{
		ID:   "u_1",
		Name: "Ada Lovelace",
	}
	s.users["cookie-456"] = User{
		ID:   "u_2",
		Name: "Alan Turing",
	}

	// Initialize profiles (user ID -> profile object)
	s.profiles["u_1"] = Profile{
		ID:  "u_1",
		Bio: "Mathematician & first programmer",
	}
	s.profiles["u_2"] = Profile{
		ID:  "u_2",
		Bio: "Mathematician & computer science pioneer",
	}

	// Initialize notifications (user ID -> array of notifications)
	s.notifications["u_1"] = []string{
		"Welcome to jsrpc!",
		"You have 2 new followers",
	}
	s.notifications["u_2"] = []string{
		"New feature: pipelining!",
		"Security tips for your account",
	}
}

func (s *UserServer) authenticate(args json.RawMessage) (interface{}, error) {
	// Extract session token from arguments
	var sessionToken string

	// Try to parse as array first
	var argArray []string
	if err := json.Unmarshal(args, &argArray); err == nil && len(argArray) > 0 {
		sessionToken = argArray[0]
	} else {
		// Try to parse as string
		if err := json.Unmarshal(args, &sessionToken); err != nil {
			return nil, err
		}
	}

	// Look up user by session token
	user, exists := s.users[sessionToken]
	if !exists {
		return nil, fmt.Errorf("invalid session")
	}

	return user, nil
}

func (s *UserServer) getUserProfile(args json.RawMessage) (interface{}, error) {
	// Extract user ID from arguments
	var userID string

	// Try to parse as array first
	var argArray []string
	if err := json.Unmarshal(args, &argArray); err == nil && len(argArray) > 0 {
		userID = argArray[0]
	} else {
		// Try to parse as string
		if err := json.Unmarshal(args, &userID); err != nil {
			return nil, err
		}
	}

	// Look up profile by user ID
	profile, exists := s.profiles[userID]
	if !exists {
		return nil, fmt.Errorf("no such user")
	}

	return profile, nil
}

func (s *UserServer) getNotifications(args json.RawMessage) (interface{}, error) {
	// Extract user ID from arguments
	var userID string

	// Try to parse as array first
	var argArray []string
	if err := json.Unmarshal(args, &argArray); err == nil && len(argArray) > 0 {
		userID = argArray[0]
	} else {
		// Try to parse as string
		if err := json.Unmarshal(args, &userID); err != nil {
			return nil, err
		}
	}

	// Look up notifications by user ID
	notifications, exists := s.notifications[userID]
	if !exists {
		return []string{}, nil // Return empty array if no notifications
	}

	return notifications, nil
}

func main() {
	// Default to serving static files from the examples/static directory
	staticPath := "../static"
	if len(os.Args) >= 2 {
		staticPath = os.Args[1]
	}

	port := ":8000"

	// Create Echo server with middleware
	e := gocapnweb.SetupEchoServer()

	// Setup RPC endpoint
	server := NewUserServer()
	gocapnweb.SetupRpcEndpoint(e, "/rpc", server)

	// Setup static file endpoint
	gocapnweb.SetupFileEndpoint(e, "/static", staticPath)

	log.Printf("🚀 Batch Pipelining Go Server (Echo) starting on port %s", port)
	log.Printf("📁 Static files served from: %s", staticPath)
	log.Printf("🔌 HTTP Batch RPC endpoint: http://localhost%s/rpc", port)
	log.Printf("📄 Static files: http://localhost%s/static/", port)
	log.Printf("🌐 Demo URL: http://localhost%s/static/batch-pipelining/", port)
	log.Println()
	log.Println("Sample data:")
	log.Println("  Session tokens: cookie-123, cookie-456")
	log.Println("  Users: u_1 (Ada Lovelace), u_2 (Alan Turing)")

	if err := e.Start(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
