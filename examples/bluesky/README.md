# Bluesky Feed Reader Example

This example demonstrates how to use Cap'n Web RPC with AT Protocol's XRPC endpoints to build a Bluesky profile and feed viewer. It showcases:

- **AT Protocol Integration**: Bridging Cap'n Web RPC to Bluesky's XRPC API
- **Batch Pipelining**: Fetching profile and feed data in a single HTTP request
- **External API Calls**: Real-world pattern for integrating third-party APIs
- **Performance Comparison**: Toggle between batched and sequential modes

## Features

- Fetch Bluesky profiles by handle (e.g., `bsky.app`, `alice.bsky.social`)
- Display user info: avatar, bio, follower/following counts
- Show recent posts with engagement metrics
- Compare batched vs sequential request performance
- Beautiful, responsive UI with Svelte

## Running the Example

### Start the Go Backend

```bash
cd examples/bluesky
go mod tidy
go run main.go
```

The server will start on `http://localhost:8000`

### Start the Frontend

In a new terminal:

```bash
cd examples/bluesky/static
npm install
npm run dev
```

The Svelte dev server will start on `http://localhost:3000`

## How It Works

### Backend (Go)

The Go server implements two RPC methods:

1. **`getProfile(handle)`** - Fetches profile data from `app.bsky.actor.getProfile`
2. **`getFeed(handle, limit)`** - Fetches posts from `app.bsky.feed.getAuthorFeed`

Both methods call the public Bluesky API at `https://public.api.bsky.app/xrpc/`

### Frontend (Svelte)

The frontend uses the `capnweb` JavaScript library to make RPC calls:

**Batched Mode** (default):
```javascript
const [profile, feed] = await Promise.all([
  api.getProfile(handle),
  api.getFeed(handle, 10),
]);
```

Both calls are sent in a **single HTTP request** using Cap'n Web RPC pipelining!

**Sequential Mode**:
```javascript
const profile = await api.getProfile(handle);  // Request 1
const feed = await api.getFeed(handle, 10);    // Request 2
```

Two separate HTTP requests, demonstrating the performance difference.

## AT Protocol / XRPC

This example integrates with Bluesky's AT Protocol using their public XRPC endpoints:

- `app.bsky.actor.getProfile` - Get user profile information
- `app.bsky.feed.getAuthorFeed` - Get a user's posts

The Go backend acts as a bridge, translating Cap'n Web RPC calls to XRPC HTTP requests.

## Try These Handles

- `bsky.app` - Official Bluesky account
- `jay.bsky.team` - Jay Graber (Bluesky CEO)
- `pfrazee.com` - Paul Frazee (Bluesky engineer)

## Testing with curl

Test the backend directly:

```bash
# Get profile
curl -X POST http://localhost:8000/rpc \
  -d '["push",["pipeline",1,["getProfile"],["bsky.app"]]]'

curl -X POST http://localhost:8000/rpc \
  -d '["pull",1]'

# Get feed
curl -X POST http://localhost:8000/rpc \
  -d '["push",["pipeline",2,["getFeed"],["bsky.app", 5]]]'

curl -X POST http://localhost:8000/rpc \
  -d '["pull",2]'
```

## Performance Benefits

Batched mode typically shows **40-60% faster** load times compared to sequential mode by:

1. Reducing network round trips (1 request vs 2)
2. Eliminating network latency between requests
3. Pipelining dependent operations efficiently

## Architecture

```
Browser (Svelte)
    ↓ Cap'n Web RPC (batched)
Go Server (localhost:8000)
    ↓ HTTPS
AT Protocol API (public.api.bsky.app)
    ↓ XRPC
Bluesky Network
```

## Error Handling

The example handles:

- Invalid or non-existent handles
- Network failures
- API rate limits
- Malformed responses

## Learn More

- [Cap'n Web RPC Spec](https://github.com/cloudflare/capnweb)
- [AT Protocol Docs](https://atproto.com)
- [Bluesky XRPC API](https://docs.bsky.app)

