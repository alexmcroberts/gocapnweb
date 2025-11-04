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
- Go server running (see the Go example in `../../helloworld/`)

### Installation

1. Install dependencies:
```bash
npm install
```

### Development

1. Start the Go server (from the `../../helloworld/` directory):
```bash
cd ../../helloworld/
go run main.go
```

2. In a new terminal, start the Svelte development server:
```bash
npm run dev
```

3. Open your browser to `http://localhost:3000`

### Production Build

To build for production:

```bash
npm run build
```

The built files will be in the `dist/` directory.

To preview the production build:

```bash
npm run preview
```

## What's Different from the Original

- **Reactive State**: Uses Svelte's reactive variables instead of DOM manipulation
- **Component-based**: All UI logic is contained in a single Svelte component
- **Better UX**: Loading states, disabled buttons during requests, and smoother interactions
- **Modern Build**: Uses Vite for fast builds and hot module replacement
- **Cleaner Code**: More maintainable and readable code structure

## File Structure

```
├── src/
│   ├── App.svelte          # Main Svelte component
│   └── main.js             # Entry point
├── dist/                   # Built files (after npm run build)
├── index.html              # HTML template
├── package.json            # Dependencies and scripts
├── vite.config.js          # Vite configuration
└── svelte.config.js        # Svelte configuration
```

## API Usage

The demo connects to the Go server's WebSocket endpoint at `ws://127.0.0.1:8000/api` and calls the `hello` method with a name parameter.

Example RPC call:
```javascript
const result = await api.hello("World");
// Returns: "Hello, World!"
```
