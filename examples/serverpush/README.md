# Server Push Example

This example demonstrates Server Push technology using the Cap'n Web RPC protocol with real-time data streaming capabilities.

## What This Demo Shows

- **System Metrics Monitoring**: CPU, memory, and network I/O metrics updated in real-time  
- **Subscription-based Data Feeds**: Client subscribes to specific data streams
- **Dynamic Data Visualization**: Real-time charts and progress bars
- **Polling-based Server Push**: Efficient polling mechanism simulating server push

## Architecture

Since the Cap'n Web RPC protocol is primarily request-response based, this example implements Server Push using a smart polling approach:

1. **Subscription Management**: Clients subscribe to data streams and receive subscription IDs
2. **Data Buffering**: Server buffers updates for each subscription in memory
3. **Polling Interface**: Clients poll for updates using their subscription ID
4. **Real-time Data Generation**: Background goroutines generate realistic fake data

## Running the Example

1. **Start the Server**:
   ```bash
   go run main.go
   ```

2. In a new terminal, start the Svelte development server:
   ```bash
   cd static
   npm install
   npm run dev
   ```

3. Open your browser to `http://localhost:3000`

4. **Test the API** (optional):
   ```bash
   # Subscribe to system metrics
   echo '["push",["pipeline",1,["subscribeSystemMetrics"],[]]]
   ["pull",1]' | curl -X POST http://localhost:8000/api -H "Content-Type: text/plain" --data-binary @-
   ```
This example serves as a foundation for implementing real-time data streaming in applications that need to push updates to clients while working within the constraints of request-response protocols.
