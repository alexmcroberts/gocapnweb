package main

import (
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"runtime/metrics"
	"sync"
	"time"

	"github.com/gocapnweb"
)

// SystemMetrics represents system performance metrics
type SystemMetrics struct {
	CPUUsage    float64 `json:"cpuUsage"`
	MemoryUsage float64 `json:"memoryUsage"`
	NetworkIO   float64 `json:"networkIO"`
	Timestamp   int64   `json:"timestamp"`
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
			"cpuUsage":    latest.CPUUsage,
			"memoryUsage": latest.MemoryUsage,
			"networkIO":   latest.NetworkIO,
			"timestamp":   latest.Timestamp,
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

// collectRealSystemMetrics collects actual system metrics using runtime/metrics
// This replaces the previous fake metrics generation with real Go runtime data:
// - CPUUsage: Based on GC CPU time ratio and goroutine activity
// - MemoryUsage: Percentage of allocated memory vs system memory (from runtime.MemStats)
// - NetworkIO: Uses GC cycles as a proxy for system activity (scaled to 0-100)
func (s *MetricsServer) collectRealSystemMetrics() SystemMetrics {
	// Get available metrics
	descs := metrics.All()

	// Create samples for the metrics we want
	samples := make([]metrics.Sample, 0, len(descs))

	// Metrics we're interested in
	metricNames := map[string]bool{
		"/cpu/classes/gc/total:cpu-seconds": true,
		"/cpu/classes/total:cpu-seconds":    true,
		"/gc/cycles/total:gc-cycles":        true,
	}

	// Build samples for metrics we want
	for _, desc := range descs {
		if metricNames[desc.Name] {
			samples = append(samples, metrics.Sample{Name: desc.Name})
		}
	}

	// Read the metrics
	metrics.Read(samples)

	// Process the results
	var totalCPU, gcCPU, gcCycles float64

	for _, sample := range samples {
		switch sample.Name {
		case "/cpu/classes/total:cpu-seconds":
			if sample.Value.Kind() == metrics.KindFloat64 {
				totalCPU = sample.Value.Float64()
			}
		case "/cpu/classes/gc/total:cpu-seconds":
			if sample.Value.Kind() == metrics.KindFloat64 {
				gcCPU = sample.Value.Float64()
			}
		case "/gc/cycles/total:gc-cycles":
			if sample.Value.Kind() == metrics.KindUint64 {
				gcCycles = float64(sample.Value.Uint64())
			}
		}
	}

	// Get additional runtime stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Calculate CPU usage as a percentage (approximate based on GC overhead)
	cpuUsage := 0.0
	if totalCPU > 0 {
		// Use GC CPU time as a rough indicator of system load
		// This is not perfect but gives us a real metric
		cpuUsage = math.Min((gcCPU/totalCPU)*100, 100.0)
		if cpuUsage < 1.0 {
			cpuUsage = float64(runtime.NumGoroutine()) / float64(runtime.GOMAXPROCS(0)) * 10.0
		}
	}

	// Calculate memory usage as a percentage of allocated vs system memory
	memoryUsage := 0.0
	if memStats.Sys > 0 {
		memoryUsage = (float64(memStats.Alloc) / float64(memStats.Sys)) * 100
	}

	// Use GC cycles as a proxy for "network IO" activity
	// This represents system activity which is more meaningful than fake network data
	networkIO := math.Min(gcCycles/1000.0, 100.0) // Scale down and cap at 100

	return SystemMetrics{
		CPUUsage:    math.Round(cpuUsage*100) / 100,
		MemoryUsage: math.Round(memoryUsage*100) / 100,
		NetworkIO:   math.Round(networkIO*100) / 100,
		Timestamp:   time.Now().Unix(),
	}
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
	staticPath := "../static"
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
	log.Printf("📄 Static files: http://localhost%s/static/", port)
	log.Printf("🌐 Demo URL: http://localhost%s/static/serverpush/", port)
	log.Println()
	log.Println("Server Push Features:")
	log.Println("  💻 Live system metrics streaming")
	log.Println("  🔄 WebSocket-based push notifications")

	if err := e.Start(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
