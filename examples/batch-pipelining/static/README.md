# Batch Pipelining Demo - Cap'n Web RPC Svelte

This is a Svelte version of the Batch Pipelining demo that showcases the performance benefits of batching multiple dependent RPC calls into a single HTTP request using Cap'n Web RPC protocol.

## Features

- **Batch Pipelining**: Multiple dependent RPC calls batched into a single HTTP request
- **Performance Comparison**: Side-by-side comparison of pipelined vs sequential execution
- **Pipeline References**: Results from one call can be used as input to subsequent calls
- **Svelte Frontend**: Modern reactive UI with component-based architecture
- **Reactive State Management**: Centralized state using Svelte stores
- **Real-time Performance Metrics**: Live timing and request count tracking
- **Hot Module Replacement**: Fast development with Vite
- **Network Simulation**: Configurable RTT and jitter simulation

## Architecture

### Frontend (Svelte)
- **App.svelte**: Main application component
- **ConfigCard.svelte**: Configuration display component
- **ResultsCard.svelte**: Individual result display with expandable details
- **SummaryCard.svelte**: Performance comparison summary
- **demoStore.js**: Centralized state management and demo logic

### Backend (Go)
- User authentication simulation with session tokens
- User profile and notification retrieval
- Batch RPC processing with pipeline reference support
- HTTP-based RPC endpoint (not WebSocket like other examples)

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
- **Centralized State**: All demo state managed through Svelte stores
- **Better UX**: Loading states, disabled buttons during execution, and smoother interactions
- **Modern Build System**: Uses Vite for fast builds and hot module replacement
- **Enhanced Error Handling**: Better error states and user feedback
- **Reset Functionality**: Ability to reset and re-run the demo

## File Structure

```
├── src/
│   ├── components/
│   │   ├── ConfigCard.svelte       # Configuration display
│   │   ├── ResultsCard.svelte      # Individual result display
│   │   └── SummaryCard.svelte      # Performance summary
│   ├── stores/
│   │   └── demoStore.js            # Demo logic and state management
│   ├── App.svelte                  # Main application component
│   └── main.js                     # Entry point
├── dist/                           # Built files (after npm run build)
├── index.html                      # HTML template
├── package.json                    # Dependencies and scripts
├── vite.config.js                  # Vite configuration
├── svelte.config.js                # Svelte configuration
└── client.mjs.backup               # Backup of original client code
```

## Demo Flow

The demo demonstrates the performance benefits of pipelining by:

1. **Pipelined Execution**: 
   - Creates a single batch RPC session
   - Makes three dependent calls: authenticate → getUserProfile → getNotifications
   - All calls are batched into a single HTTP request
   - Pipeline references allow dependent calls without waiting

2. **Sequential Execution**:
   - Creates separate RPC sessions for each call
   - Makes three separate HTTP requests in sequence
   - Each call waits for the previous one to complete

3. **Performance Comparison**:
   - Measures execution time and HTTP request count
   - Shows dramatic latency reduction with pipelining
   - Demonstrates the benefits in high-latency network scenarios

## Configuration

The demo supports URL parameters for configuration:

- `rpc_url`: RPC endpoint URL (default: `http://localhost:8000/rpc`)
- `rtt`: Simulated round-trip time in ms (default: 120)
- `jitter`: RTT jitter in ms (default: 40)
- `autorun`: Auto-run demo on load (default: false)

Example: `http://localhost:3000?rtt=200&jitter=50&autorun=true`

## Sample Data

The Go server includes sample data:

- **Session Tokens**: `cookie-123`, `cookie-456`
- **Users**: 
  - `u_1`: Ada Lovelace (Mathematician & first programmer)
  - `u_2`: Alan Turing (Mathematician & computer science pioneer)
- **Notifications**: Welcome messages and feature updates

## Key Benefits Demonstrated

- **Reduced Latency**: Single round trip vs multiple round trips
- **Pipeline References**: Dependent calls without intermediate waits
- **Network Efficiency**: Fewer HTTP connections and requests
- **Scalability**: Better performance in high-latency environments
