# Batch Pipelining Example

This example demonstrates the power of Cap'n Web RPC's batch pipelining feature by showing how multiple dependent RPC calls can be executed in a single HTTP request.

## What This Demo Shows

- **Batch Pipelining**: Multiple dependent RPC calls in a single HTTP request
- **Pipeline References**: Using the result of one call as input to subsequent calls
- **Performance Comparison**: Side-by-side comparison of pipelined vs sequential execution
- **Network Latency Simulation**: Configurable simulated RTT to demonstrate real-world benefits
- **Reactive UI**: Built with Svelte 5 for smooth, reactive user experience

## Running the Example

1. **Start the Go Server**:
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

## The Problem

Traditional RPC requires multiple round trips for dependent operations:

```javascript
// Sequential: 3 separate HTTP requests (3 round trips)
const user = await api.authenticate('cookie-123');     // Round trip 1
const profile = await api.getUserProfile(user.id);     // Round trip 2  
const notifications = await api.getNotifications(user.id); // Round trip 3
```

With typical network latency (120ms RTT), this takes ~360ms just for network overhead!

## The Solution: Pipeline References

Cap'n Web RPC's pipelining allows dependent calls to be batched:

```javascript
// Pipelined: All 3 calls in 1 HTTP request (1 round trip)
const api = newHttpBatchRpcSession(RPC_URL);
const user = api.authenticate('cookie-123');
const profile = api.getUserProfile(user.id);           // References user.id before it resolves!
const notifications = api.getNotifications(user.id);   // References user.id before it resolves!

const [u, p, n] = await Promise.all([user, profile, notifications]);
```

All three calls are sent together in a single HTTP request. The server resolves them in dependency order and returns all results at once!

## How It Works

### Backend (Go)

The Go server implements three RPC methods:

1. **`authenticate(sessionToken)`** - Returns a User object with ID and name
2. **`getUserProfile(userID)`** - Returns profile data for a user
3. **`getNotifications(userID)`** - Returns an array of notifications for a user

Sample data:
- Session tokens: `cookie-123`, `cookie-456`
- Users: `u_1` (Ada Lovelace), `u_2` (Alan Turing)

### Frontend (Svelte)

**Pipelined Execution** (1 HTTP request):
```javascript
const api = newHttpBatchRpcSession(RPC_URL);
const user = api.authenticate('cookie-123');
const profile = api.getUserProfile(user.id);
const notifications = api.getNotifications(user.id);

const [u, p, n] = await Promise.all([user, profile, notifications]);
```

**Sequential Execution** (3 HTTP requests):
```javascript
const api1 = newHttpBatchRpcSession(RPC_URL);
const u = await api1.authenticate('cookie-123');

const api2 = newHttpBatchRpcSession(RPC_URL);
const p = await api2.getUserProfile(u.id);

const api3 = newHttpBatchRpcSession(RPC_URL);
const n = await api3.getNotifications(u.id);
```

## Performance Benefits

With default simulated latency (120ms RTT ± 40ms jitter):

- **Pipelined**: ~1 round trip = ~120-160ms
- **Sequential**: ~3 round trips = ~360-480ms

**Pipelined execution is typically 60-70% faster!**

The benefits increase with:
- Higher network latency
- More dependent operations
- Geographic distance between client and server

## Testing with curl

Test the backend directly:

```bash
# Authenticate
curl -X POST http://localhost:8000/rpc \
  -d '["push",["pipeline",1,["authenticate"],["cookie-123"]]]'

curl -X POST http://localhost:8000/rpc \
  -d '["pull",1]'

# Get profile (using user ID u_1)
curl -X POST http://localhost:8000/rpc \
  -d '["push",["pipeline",2,["getUserProfile"],["u_1"]]]'

curl -X POST http://localhost:8000/rpc \
  -d '["pull",2]'

# Get notifications
curl -X POST http://localhost:8000/rpc \
  -d '["push",["pipeline",3,["getNotifications"],["u_1"]]]'

curl -X POST http://localhost:8000/rpc \
  -d '["pull",3]'
```