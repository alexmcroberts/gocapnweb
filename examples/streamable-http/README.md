# Streamable HTTP – AI typewriter

This example demonstrates **streaming HTTP responses**: the server sends the response body in chunks using chunked transfer encoding and flushes each chunk immediately. The client uses `fetch()` and reads `response.body` with a `ReadableStream`, updating the UI in real time. No WebSocket—plain HTTP.

## What it demonstrates

- **Streaming response**: Server writes tokens (simulated) one at a time and flushes after each write.
- **Chunked transfer encoding**: Response has no `Content-Length`; the server uses `http.Flusher` to send chunks as they’re ready.
- **Svelte 5**: Reactive state (`$state`) is updated as the stream is read; the typewriter effect is visible in the UI.
- **Fetch + ReadableStream**: Client uses `response.body.getReader()` and `TextDecoder` to consume the stream incrementally.

## Running

**Terminal 1 – Go server**

```bash
cd streamable-http
go run main.go
```

**Terminal 2 – Svelte dev server**

```bash
cd static
npm install
npm run dev
```

Open http://localhost:3000. Enter a prompt (or leave the default), click **Generate**, and watch the text appear token-by-token.

## API

- **POST /api/generate-stream**  
  - Request body: plain text (optional prompt; first line used).  
  - Response: `text/plain`, streamed. Server sends simulated tokens with ~80 ms delay between chunks; each write is flushed immediately.

## Frontend

- **Svelte 5** with `$state` for prompt, streamed text, and status.
- **Vite** dev server proxies `/api` to the Go server on port 8000.
- **Generate** starts a streaming request; **Stop** aborts it; **Reset** clears the output.
