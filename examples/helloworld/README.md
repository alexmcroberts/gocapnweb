# Hello World - Cap'n Web RPC Svelte Demo

This is a Svelte version of the Hello World demo that shows WebSocket RPC communication between a Svelte frontend and a Go server using the Cap'n Web protocol.

## Features

- **Svelte Frontend**: Modern reactive UI built with Svelte
- **WebSocket RPC**: Real-time bidirectional communication
- **Hot Module Replacement**: Fast development with Vite
- **TypeScript Support**: Ready for TypeScript if needed

## Setup and Running

### Prerequisites

- Node.js (v16 or higher)
– Golang v1.21 or higher

### Running the Example

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