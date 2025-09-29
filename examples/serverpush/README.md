# Server Push Example

This example demonstrates Server Push technology using the Cap'n Web RPC protocol with real-time data streaming capabilities.

## What This Demo Shows

- 💻 **System Metrics Monitoring**: CPU, memory, and network I/O metrics updated in real-time  
- 🔄 **Subscription-based Data Feeds**: Client subscribes to specific data streams
- 📊 **Dynamic Data Visualization**: Real-time charts and progress bars
- ⚡ **Polling-based Server Push**: Efficient polling mechanism simulating server push

## Architecture

Since the Cap'n Web RPC protocol is primarily request-response based, this example implements Server Push using a smart polling approach:

1. **Subscription Management**: Clients subscribe to data streams and receive subscription IDs
2. **Data Buffering**: Server buffers updates for each subscription in memory
3. **Polling Interface**: Clients poll for updates using their subscription ID
4. **Real-time Data Generation**: Background goroutines generate realistic fake data

## Running the Example

1. **Start the Server**:
   ```bash
   cd examples/serverpush
   go run main.go
   # or build and run:
   go build && ./serverpush
   ```

2. **Open the Demo**:
   ```
   http://localhost:8000/static/serverpush/
   ```

3. **Test the API** (optional):
   ```bash
   # Subscribe to system metrics
   echo '["push",["pipeline",1,["subscribeSystemMetrics"],[]]]
   ["pull",1]' | curl -X POST http://localhost:8000/api -H "Content-Type: text/plain" --data-binary @-
   ```

## Available RPC Methods

### `subscribeSystemMetrics()`
- **Returns**: `{subscriptionId, message, status, pollInterval}`
- **Description**: Subscribe to real-time system metrics

### `pollMetricsUpdates(subscriptionId)`
- **Parameters**: `subscriptionId` (string)
- **Returns**: `{subscriptionId, updates, hasMore, timestamp}`
- **Description**: Poll for buffered system metrics updates

### `unsubscribe(subscriptionId)`
- **Parameters**: `subscriptionId` (string)
- **Returns**: `{subscriptionId, message, status}`
- **Description**: Unsubscribe from a data stream

## Data Formats

### System Metrics Update
```json
{
  "cpuUsage": 45.2,
  "memoryUsage": 62.8,
  "networkIO": 23.4,
  "timestamp": 1640995200
}
```

## Features Demonstrated

- **WebSocket RPC Connection**: Persistent connection for low-latency communication
- **Subscription Management**: Server tracks client subscriptions with unique IDs
- **Data Buffering**: Efficient buffering prevents data loss during polling gaps
- **Real-time Visualization**: Dynamic charts, tables, and progress indicators
- **Error Handling**: Graceful handling of connection issues and invalid subscriptions
- **Resource Management**: Automatic cleanup of subscriptions and polling intervals

## Technical Implementation

- **Backend**: Go with Echo framework and Gorilla WebSocket
- **Frontend**: Vanilla JavaScript with modern ES6 modules
- **RPC Protocol**: Cap'n Web RPC over WebSocket
- **Data Generation**: Background goroutines with realistic price/metrics simulation
- **Concurrency**: Thread-safe subscription management with sync.RWMutex

This example serves as a foundation for implementing real-time data streaming in applications that need to push updates to clients while working within the constraints of request-response protocols.
