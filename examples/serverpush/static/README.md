# System Metrics Demo - Cap'n Web RPC Svelte

This is a Svelte version of the System Metrics demo that showcases real-time server push functionality using WebSocket RPC communication between a Svelte frontend and a Go server with the Cap'n Web protocol.

## Features

- **Real-time System Metrics**: Live CPU percentage, Disk usage, and Network I/O monitoring
- **Server Push Technology**: Subscription-based data feeds with automatic polling
- **Svelte Frontend**: Modern reactive UI with component-based architecture
- **Reactive State Management**: Centralized state using Svelte stores
- **Real-time Data Visualization**: Progress bars and live updating metrics
- **Hot Module Replacement**: Fast development with Vite
- **Accurate System Monitoring**: Uses gopsutil for precise system metrics collection

## Architecture

### Frontend (Svelte)
- **App.svelte**: Main application component
- **MetricCard.svelte**: Reusable component for displaying individual metrics
- **ConnectionStatus.svelte**: Component for showing connection state
- **appStore.js**: Centralized state management using Svelte stores

### Backend (Go)
- Real-time system metrics collection using `gopsutil` library
- Subscription-based polling system for efficient data delivery
- WebSocket RPC endpoints for client communication
- Accurate CPU percentage, disk usage, and network I/O monitoring

## Setup and Running

### Prerequisites

- Node.js (v16 or higher)
- Go server running (see the Go example in `../`)

### Installation

1. Install dependencies:
```bash
npm install
```

### Development

1. Start the Go server (from the parent directory):
```bash
cd ../
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

- **Reactive Architecture**: Uses Svelte's reactive variables and stores instead of DOM manipulation
- **Component-based Design**: Modular components for better maintainability
- **Centralized State**: All application state managed through Svelte stores
- **Better UX**: Loading states, disabled buttons during requests, and smoother interactions
- **Modern Build System**: Uses Vite for fast builds and hot module replacement
- **Improved Error Handling**: Better error states and user feedback

## File Structure

```
├── src/
│   ├── components/
│   │   ├── MetricCard.svelte      # Reusable metric display component
│   │   └── ConnectionStatus.svelte # Connection status indicator
│   ├── stores/
│   │   └── appStore.js            # Centralized state management
│   ├── App.svelte                 # Main application component
│   └── main.js                    # Entry point
├── dist/                          # Built files (after npm run build)
├── index.html                     # HTML template
├── package.json                   # Dependencies and scripts
├── vite.config.js                 # Vite configuration
├── svelte.config.js               # Svelte configuration
└── client.mjs.backup              # Backup of original client code
```

## API Usage

The demo connects to the Go server's WebSocket endpoint at `ws://127.0.0.1:8000/api` and uses the following RPC methods:

### Subscribe to System Metrics
```javascript
const response = await api.subscribeSystemMetrics();
// Returns: { subscriptionId, message, status, pollInterval }
```

### Poll for Metrics Updates
```javascript
const response = await api.pollMetricsUpdates(subscriptionId);
// Returns: { subscriptionId, latestMetrics, hasData, updateCount, timestamp }
```

### Unsubscribe
```javascript
const response = await api.unsubscribe(subscriptionId);
// Returns: { subscriptionId, message, status }
```

## System Metrics

The demo displays real-time system metrics collected from the Go server using gopsutil:

- **CPU Percentage**: Real CPU usage percentage across all cores
- **Disk Usage**: Disk usage percentage for the root filesystem (/)
- **Network I/O**: Network activity (scaled representation of total bytes transferred)

All metrics are updated every second and displayed with smooth animations and progress bars.
