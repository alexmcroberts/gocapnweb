// Import the capnweb client library
import { newWebSocketRpcSession } from 'https://unpkg.com/capnweb@latest/dist/index.js';

let api = null;
let isConnected = false;
let metricsSubscriptionId = null;
let metricsPollInterval = null;

// Initialize the WebSocket connection
async function connectToServer() {
    try {
        updateConnectionStatus('connecting', '🔄 Connecting to server...');
        
        // Connect to our Go server's WebSocket endpoint
        api = newWebSocketRpcSession("ws://127.0.0.1:8000/api");
        
        // Test the connection by subscribing and immediately unsubscribing
        const testResponse = await api.subscribeSystemMetrics();
        await api.unsubscribe(testResponse.subscriptionId);
        isConnected = true;
        
        updateConnectionStatus('connected', '✅ Connected to Go server successfully!');
        updateButtonStates();
        
        // Setup periodic connection check
        setupConnectionMonitoring();
        
        return true;
    } catch (error) {
        updateConnectionStatus('disconnected', `❌ Connection failed: ${error.message}`);
        isConnected = false;
        updateButtonStates();
        return false;
    }
}

// Subscribe to real-time system metrics
async function subscribeToMetrics() {
    if (!isConnected) {
        alert('Please connect to the server first');
        return;
    }
    
    try {
        const response = await api.subscribeSystemMetrics();
        metricsSubscriptionId = response.subscriptionId;
        
        document.getElementById('metricsSubscriptionStatus').textContent = 
            `✅ Subscribed (ID: ${metricsSubscriptionId})`;
        document.getElementById('metricsBtn').disabled = true;
        
        // Start polling for metrics updates
        if (response.pollInterval) {
            startMetricsPolling(response.pollInterval);
        }
        
        console.log('Subscribed to system metrics:', response);
    } catch (error) {
        console.error('Failed to subscribe to system metrics:', error);
        alert(`Failed to subscribe: ${error.message}`);
    }
}

// Unsubscribe from metrics feed
async function unsubscribeAll() {
    try {
        if (metricsSubscriptionId) {
            await api.unsubscribe(metricsSubscriptionId);
            stopMetricsPolling();
            metricsSubscriptionId = null;
            document.getElementById('metricsSubscriptionStatus').textContent = 'Not subscribed';
            document.getElementById('metricsBtn').disabled = false;
        }
        
        console.log('Unsubscribed from metrics feed');
    } catch (error) {
        console.error('Failed to unsubscribe:', error);
    }
}

// Removed getCurrentPrices - focusing only on system metrics

// Update connection status display
function updateConnectionStatus(status, message) {
    const statusElement = document.getElementById('connectionStatus');
    const statusText = document.getElementById('statusText');
    
    statusElement.className = `status ${status}`;
    statusText.textContent = message;
}

// Update button states based on connection status
function updateButtonStates() {
    const buttons = {
        connectBtn: document.getElementById('connectBtn'),
        metricsBtn: document.getElementById('metricsBtn'),
        unsubscribeBtn: document.getElementById('unsubscribeBtn')
    };
    
    buttons.connectBtn.disabled = isConnected;
    buttons.metricsBtn.disabled = !isConnected || metricsSubscriptionId !== null;
    buttons.unsubscribeBtn.disabled = !isConnected || !metricsSubscriptionId;
}

// Update system metrics display
function updateMetricsDisplay(metricsData) {
    // Update CPU
    document.getElementById('cpuValue').textContent = `${metricsData.cpuUsage.toFixed(1)}%`;
    document.getElementById('cpuProgress').style.width = `${metricsData.cpuUsage}%`;
    
    // Update Memory
    document.getElementById('memoryValue').textContent = `${metricsData.memoryUsage.toFixed(1)}%`;
    document.getElementById('memoryProgress').style.width = `${metricsData.memoryUsage}%`;
    
    // Update Network
    document.getElementById('networkValue').textContent = `${metricsData.networkIO.toFixed(1)} MB/s`;
    document.getElementById('networkProgress').style.width = `${metricsData.networkIO}%`;
    
    // Update timestamp
    const timestamp = new Date(metricsData.timestamp * 1000).toLocaleTimeString();
    document.getElementById('metricsTimestamp').textContent = timestamp;
}

// Setup connection monitoring
function setupConnectionMonitoring() {
    // Note: In a real implementation, you would set up WebSocket event listeners
    // to handle incoming pushed data. For now, we'll simulate this with polling
    // since the current Cap'n Web RPC protocol doesn't expose raw WebSocket events
    
    // This is a simplified approach - in practice, you'd want to enhance
    // the Cap'n Web library to support server-initiated messages
    console.log('Connection monitoring setup (simplified implementation)');
}

// Make functions available globally for HTML onclick handlers
window.connectToServer = connectToServer;
window.subscribeToMetrics = subscribeToMetrics;
window.unsubscribeAll = unsubscribeAll;

// Auto-connect when the page loads
document.addEventListener('DOMContentLoaded', async () => {
    console.log('🚀 Server Push Demo loaded');
    await connectToServer();
});

// Handle page unload - cleanup subscriptions
window.addEventListener('beforeunload', async () => {
    if (isConnected) {
        await unsubscribeAll();
    }
});

// Simulate receiving pushed data (in a real implementation, this would come through WebSocket)
// This is a demonstration of what the UI would look like with real data
function simulateIncomingData() {
    // This function would be called by the WebSocket message handler in a real implementation
    console.log('Note: In a full implementation, this would be replaced by WebSocket push events');
}

function startMetricsPolling(intervalMs) {
    console.log('startMetricsPolling called with interval:', intervalMs, 'metricsSubscriptionId:', metricsSubscriptionId);
    
    if (metricsPollInterval) {
        console.log('Clearing existing metrics poll interval');
        clearInterval(metricsPollInterval);
    }
    
    metricsPollInterval = setInterval(async () => {
        console.log('Metrics poll interval fired. metricsSubscriptionId:', metricsSubscriptionId, 'isConnected:', isConnected);
        
        // More strict check - ensure metricsSubscriptionId is a valid string
        if (!metricsSubscriptionId || typeof metricsSubscriptionId !== 'string' || metricsSubscriptionId === 'null' || !isConnected) {
            console.log('Stopping metrics polling due to invalid state');
            stopMetricsPolling();
            return;
        }
        
        try {
            console.log('About to poll metrics updates with ID:', metricsSubscriptionId);
            const response = await api.pollMetricsUpdates(metricsSubscriptionId);
            console.log('Metrics poll response:', response);
            if (response.hasData && response.latestMetrics) {
                updateMetricsDisplay(response.latestMetrics);
                console.log(`Received metrics update (${response.updateCount} updates processed)`);
            }
        } catch (error) {
            console.error('Error polling metrics updates:', error);
            console.error('metricsSubscriptionId was:', metricsSubscriptionId, 'type:', typeof metricsSubscriptionId);
            // Stop polling on error to prevent spam
            stopMetricsPolling();
        }
    }, intervalMs);
    
    console.log(`Started metrics polling every ${intervalMs}ms`);
}

function stopMetricsPolling() {
    if (metricsPollInterval) {
        clearInterval(metricsPollInterval);
        metricsPollInterval = null;
        console.log('Stopped metrics polling');
    }
}

// Development helper - simulate some data updates for demo purposes
if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
    setTimeout(() => {
        console.log('Development mode: Adding sample data visualization');
        
        // Show some sample metrics
        const sampleMetrics = {
            cpuUsage: 45.2,
            memoryUsage: 62.8,
            networkIO: 23.4,
            timestamp: Date.now() / 1000
        };
        
        updateMetricsDisplay(sampleMetrics);
    }, 2000);
}