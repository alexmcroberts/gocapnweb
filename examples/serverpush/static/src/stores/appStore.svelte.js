import { newWebSocketRpcSession } from 'capnweb';

// Application state using Svelte 5 runes
let isConnected = $state(false);
let connectionStatus = $state('disconnected');
let connectionMessage = $state('🔄 Ready to connect...');
let api = $state(null);
let metricsSubscriptionId = $state(null);
let isSubscribedToMetrics = $state(false);
let systemMetrics = $state({
  cpuPercent: 0,
  diskUsage: 0,
  networkIO: 0,
  timestamp: null
});

// Export getter functions for state access
export const getIsConnected = () => isConnected;
export const getConnectionStatus = () => connectionStatus;
export const getConnectionMessage = () => connectionMessage;
export const getApi = () => api;
export const getMetricsSubscriptionId = () => metricsSubscriptionId;
export const getIsSubscribedToMetrics = () => isSubscribedToMetrics;
export const getSystemMetrics = () => systemMetrics;

// Derived state for UI controls (private)
const canConnect = $derived(!isConnected);
const canSubscribe = $derived(isConnected && !isSubscribedToMetrics);
const canUnsubscribe = $derived(isConnected && isSubscribedToMetrics);

// Export getter functions for derived state
export const getCanConnect = () => canConnect;
export const getCanSubscribe = () => canSubscribe;
export const getCanUnsubscribe = () => canUnsubscribe;

// Polling interval reference
let metricsPollInterval = null;

// Connection functions
export async function connectToServer() {
  try {
    connectionStatus = 'connecting';
    connectionMessage = '🔄 Connecting to server...';
    
    // Connect to our Go server's WebSocket endpoint
    const apiInstance = newWebSocketRpcSession("ws://127.0.0.1:8000/api");
    
    // Test the connection by subscribing and immediately unsubscribing
    const testResponse = await apiInstance.subscribeSystemMetrics();
    await apiInstance.unsubscribe(testResponse.subscriptionId);
    
    api = apiInstance;
    isConnected = true;
    connectionStatus = 'connected';
    connectionMessage = '✅ Connected to Go server successfully!';
    
    return true;
  } catch (error) {
    connectionStatus = 'disconnected';
    connectionMessage = `❌ Connection failed: ${error.message}`;
    isConnected = false;
    api = null;
    return false;
  }
}

export async function subscribeToMetrics() {
  if (!api) {
    throw new Error('Not connected to server');
  }
  
  try {
    const response = await api.subscribeSystemMetrics();
    metricsSubscriptionId = response.subscriptionId;
    isSubscribedToMetrics = true;
    
    // Start polling for metrics updates
    if (response.pollInterval) {
      startMetricsPolling(response.pollInterval, api);
    }
    
    console.log('Subscribed to system metrics:', response);
    return response;
  } catch (error) {
    console.error('Failed to subscribe to system metrics:', error);
    throw error;
  }
}

export async function unsubscribeFromMetrics() {
  if (!api || !metricsSubscriptionId) {
    return;
  }
  
  try {
    await api.unsubscribe(metricsSubscriptionId);
    stopMetricsPolling();
    metricsSubscriptionId = null;
    isSubscribedToMetrics = false;
    
    console.log('Unsubscribed from metrics feed');
  } catch (error) {
    console.error('Failed to unsubscribe:', error);
    throw error;
  }
}

function startMetricsPolling(intervalMs, apiInstance) {
  console.log('Starting metrics polling every', intervalMs, 'ms');
  
  if (metricsPollInterval) {
    clearInterval(metricsPollInterval);
  }
  
  metricsPollInterval = setInterval(async () => {
    if (!metricsSubscriptionId || !isConnected) {
      stopMetricsPolling();
      return;
    }
    
    try {
      const response = await apiInstance.pollMetricsUpdates(metricsSubscriptionId);
      if (response.hasData && response.latestMetrics) {
        systemMetrics = response.latestMetrics;
        console.log(`Received metrics update (${response.updateCount} updates processed)`);
      }
    } catch (error) {
      console.error('Error polling metrics updates:', error);
      stopMetricsPolling();
    }
  }, intervalMs);
}

function stopMetricsPolling() {
  if (metricsPollInterval) {
    clearInterval(metricsPollInterval);
    metricsPollInterval = null;
    console.log('Stopped metrics polling');
  }
}

// Cleanup function for when the app is destroyed
export function cleanup() {
  stopMetricsPolling();
}
