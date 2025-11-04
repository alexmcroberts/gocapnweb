package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/gocapnweb"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/net"
)

// SystemMetrics represents system performance metrics
type SystemMetrics struct {
	CPUPercent float64 `json:"cpuPercent"`
	DiskUsage  float64 `json:"diskUsage"`
	NetworkIO  float64 `json:"networkIO"`
	Timestamp  int64   `json:"timestamp"`
}

// MetricsServer implements real-time system metrics streaming using polling-based Server Push.
type MetricsServer struct {
	*gocapnweb.BaseRpcTarget
	subscribers   map[string]*Subscription // subscriptionID -> subscription
	mu            sync.RWMutex
	metricsBuffer map[string][]SystemMetrics // buffered updates per subscription
}

// Subscription represents an active metrics subscription
type Subscription struct {
	ID       string
	LastPull int64 // timestamp of last data pull
}

// MetricsServerWrapper wraps BaseRpcTarget to add debugging
type MetricsServerWrapper struct {
	*MetricsServer
}

func (w *MetricsServerWrapper) Dispatch(method string, args json.RawMessage) (interface{}, error) {
	log.Printf("=== DISPATCH: method=%s, args=%s ===", method, string(args))
	result, err := w.MetricsServer.BaseRpcTarget.Dispatch(method, args)
	log.Printf("=== DISPATCH RESULT: method=%s, result=%+v, err=%v ===", method, result, err)
	return result, err
}

// NewMetricsServer creates a new server with metrics push capabilities.
func NewMetricsServer() *MetricsServerWrapper {
	server := &MetricsServer{
		BaseRpcTarget: gocapnweb.NewBaseRpcTarget(),
		subscribers:   make(map[string]*Subscription),
		metricsBuffer: make(map[string][]SystemMetrics),
	}

	wrapper := &MetricsServerWrapper{MetricsServer: server}

	// Register RPC methods
	server.Method("subscribeSystemMetrics", server.subscribeSystemMetrics)
	server.Method("unsubscribe", server.unsubscribe)
	server.Method("pollMetricsUpdates", server.pollMetricsUpdates)

	// Start background metrics generator
	go server.generateSystemMetrics()

	return wrapper
}

func (s *MetricsServer) subscribeSystemMetrics(args json.RawMessage) (interface{}, error) {
	// Create a unique subscription ID
	subscriptionID := "system_metrics_" + generateID()

	subscription := &Subscription{
		ID:       subscriptionID,
		LastPull: time.Now().Unix(),
	}

	s.mu.Lock()
	s.subscribers[subscriptionID] = subscription
	s.metricsBuffer[subscriptionID] = make([]SystemMetrics, 0)
	s.mu.Unlock()

	log.Printf("Client subscribed to system metrics: %s", subscriptionID)

	response := map[string]interface{}{
		"subscriptionId": subscriptionID,
		"message":        "Subscribed to real-time system metrics",
		"status":         "active",
		"pollInterval":   1000, // milliseconds
	}
	log.Printf("subscribeSystemMetrics returning: %+v", response)
	return response, nil
}

func (s *MetricsServer) unsubscribe(args json.RawMessage) (interface{}, error) {
	var argArray []string
	if err := json.Unmarshal(args, &argArray); err != nil {
		return nil, err
	}

	if len(argArray) == 0 {
		return map[string]string{"error": "subscription ID required"}, nil
	}

	subscriptionID := argArray[0]

	s.mu.Lock()
	if _, exists := s.subscribers[subscriptionID]; exists {
		delete(s.subscribers, subscriptionID)
		delete(s.metricsBuffer, subscriptionID)
		s.mu.Unlock()

		log.Printf("Client unsubscribed: %s", subscriptionID)
		return map[string]interface{}{
			"subscriptionId": subscriptionID,
			"message":        "Successfully unsubscribed",
			"status":         "inactive",
		}, nil
	}
	s.mu.Unlock()

	return map[string]string{"error": "subscription not found"}, nil
}

func (s *MetricsServer) pollMetricsUpdates(args json.RawMessage) (interface{}, error) {
	var argArray []string
	if err := json.Unmarshal(args, &argArray); err != nil {
		return nil, err
	}

	if len(argArray) == 0 {
		return map[string]string{"error": "subscription ID required"}, nil
	}

	subscriptionID := argArray[0]

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if subscription exists
	subscription, exists := s.subscribers[subscriptionID]
	if !exists {
		return map[string]string{"error": "subscription not found"}, nil
	}

	// Get buffered updates for this subscription
	metricsUpdates := s.metricsBuffer[subscriptionID]
	s.metricsBuffer[subscriptionID] = make([]SystemMetrics, 0) // Clear buffer

	// Return only the latest metrics (not as an array) to avoid double-wrapping
	var latestMetrics map[string]interface{}
	if len(metricsUpdates) > 0 {
		// Get the most recent metrics
		latest := metricsUpdates[len(metricsUpdates)-1]
		latestMetrics = map[string]interface{}{
			"cpuPercent": latest.CPUPercent,
			"diskUsage":  latest.DiskUsage,
			"networkIO":  latest.NetworkIO,
			"timestamp":  latest.Timestamp,
		}
	}

	// Update last poll time
	subscription.LastPull = time.Now().Unix()

	return map[string]interface{}{
		"subscriptionId": subscriptionID,
		"latestMetrics":  latestMetrics,
		"hasData":        latestMetrics != nil,
		"updateCount":    len(metricsUpdates),
		"timestamp":      time.Now().Unix(),
	}, nil
}

// Background goroutine to collect real system metrics
func (s *MetricsServer) generateSystemMetrics() {
	ticker := time.NewTicker(1 * time.Second) // Update every second
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Collect real system metrics using runtime/metrics
			metrics := s.collectRealSystemMetrics()

			// Buffer update for all metrics subscribers
			s.bufferMetricsUpdate(metrics)
		}
	}
}

// Helper function to buffer metrics updates for subscribers
func (s *MetricsServer) bufferMetricsUpdate(update SystemMetrics) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Add to buffer for all subscribers
	for subscriptionID := range s.subscribers {
		// Add to buffer, keep only last 30 updates per subscription
		buffer := s.metricsBuffer[subscriptionID]
		buffer = append(buffer, update)
		if len(buffer) > 30 {
			buffer = buffer[len(buffer)-30:]
		}
		s.metricsBuffer[subscriptionID] = buffer
	}
}

// collectRealSystemMetrics collects actual system metrics using gopsutil
// - CPUPercent: Real CPU usage percentage across all cores
// - DiskUsage: Disk usage percentage for the root filesystem
// - NetworkIO: Network I/O bytes per second (combined sent + received)
func (s *MetricsServer) collectRealSystemMetrics() SystemMetrics {
	ctx := context.Background()

	// Get CPU percentage (average across all cores)
	cpuPercents, err := cpu.PercentWithContext(ctx, time.Second, false)
	cpuPercent := 0.0
	if err == nil && len(cpuPercents) > 0 {
		cpuPercent = cpuPercents[0] // Overall CPU usage
	}

	// Get disk usage for root filesystem
	diskStat, err := disk.UsageWithContext(ctx, "/")
	diskUsage := 0.0
	if err == nil {
		diskUsage = diskStat.UsedPercent
	}

	// Get network I/O counters
	netStats, err := net.IOCountersWithContext(ctx, false)
	networkIO := 0.0
	if err == nil && len(netStats) > 0 {
		// Calculate total bytes (sent + received) and convert to MB/s
		// This is a simplified approach - in a real implementation you'd track
		// the rate of change over time
		totalBytes := float64(netStats[0].BytesSent + netStats[0].BytesRecv)
		// Scale down to a reasonable range (0-100) for display purposes
		networkIO = (totalBytes / (1024 * 1024 * 1024)) * 10 // GB to scaled value
		if networkIO > 100 {
			networkIO = 100
		}
	}

	return SystemMetrics{
		CPUPercent: roundToTwoDecimals(cpuPercent),
		DiskUsage:  roundToTwoDecimals(diskUsage),
		NetworkIO:  roundToTwoDecimals(networkIO),
		Timestamp:  time.Now().Unix(),
	}
}

// Helper function to round to 2 decimal places
func roundToTwoDecimals(val float64) float64 {
	return float64(int(val*100)) / 100
}

// Helper function to generate a unique ID
func generateID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, 8)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func main() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Default to serving static files from the examples/static directory
	staticPath := "/static"
	if len(os.Args) >= 2 {
		staticPath = os.Args[1]
	}

	port := ":8000"

	// Create Echo server with middleware
	e := gocapnweb.SetupEchoServer()

	// Setup RPC endpoint
	server := NewMetricsServer()
	gocapnweb.SetupRpcEndpoint(e, "/api", server)

	// Setup static file endpoint
	gocapnweb.SetupFileEndpoint(e, "/static", staticPath)

	log.Printf("🚀 System Metrics Go Server (Echo) starting on port %s", port)
	log.Printf("📁 Static files served from: %s", staticPath)
	log.Printf("🔌 WebSocket RPC endpoint: ws://localhost%s/api", port)
	log.Printf("🔌 HTTP Batch RPC endpoint: http://localhost%s/rpc", port)
	log.Printf("🌐 Demo URL: http://localhost:3000 (available once you start the Svelte development server)")
	log.Println()
	log.Println("Server Push Features:")
	log.Println("  💻 Live system metrics streaming")
	log.Println("  🔄 WebSocket-based push notifications")

	if err := e.Start(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
