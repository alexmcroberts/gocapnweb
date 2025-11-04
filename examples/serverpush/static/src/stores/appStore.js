import { writable, derived, get } from 'svelte/store';
import { newWebSocketRpcSession } from 'capnweb';

// Connection state
export const isConnected = writable(false);
export const connectionStatus = writable('disconnected');
export const connectionMessage = writable('🔄 Ready to connect...');

// API instance
export const api = writable(null);

// Metrics subscription state
export const metricsSubscriptionId = writable(null);
export const isSubscribedToMetrics = writable(false);

// System metrics data
export const systemMetrics = writable({
  cpuPercent: 0,
  diskUsage: 0,
  networkIO: 0,
  timestamp: null
});

// Polling interval reference
let metricsPollInterval = null;

// Derived stores for UI state
export const canConnect = derived(isConnected, $isConnected => !$isConnected);
export const canSubscribe = derived(
  [isConnected, isSubscribedToMetrics], 
  ([$isConnected, $isSubscribedToMetrics]) => $isConnected && !$isSubscribedToMetrics
);
export const canUnsubscribe = derived(
  [isConnected, isSubscribedToMetrics], 
  ([$isConnected, $isSubscribedToMetrics]) => $isConnected && $isSubscribedToMetrics
);

// Connection functions
export async function connectToServer() {
  try {
    connectionStatus.set('connecting');
    connectionMessage.set('🔄 Connecting to server...');
    
    // Connect to our Go server's WebSocket endpoint
    const apiInstance = newWebSocketRpcSession("ws://127.0.0.1:8000/api");
    
    // Test the connection by subscribing and immediately unsubscribing
    const testResponse = await apiInstance.subscribeSystemMetrics();
    await apiInstance.unsubscribe(testResponse.subscriptionId);
    
    api.set(apiInstance);
    isConnected.set(true);
    connectionStatus.set('connected');
    connectionMessage.set('✅ Connected to Go server successfully!');
    
    return true;
  } catch (error) {
    connectionStatus.set('disconnected');
    connectionMessage.set(`❌ Connection failed: ${error.message}`);
    isConnected.set(false);
    api.set(null);
    return false;
  }
}

export async function subscribeToMetrics() {
  const $api = get(api);
  
  if (!$api) {
    throw new Error('Not connected to server');
  }
  
  try {
    const response = await $api.subscribeSystemMetrics();
    metricsSubscriptionId.set(response.subscriptionId);
    isSubscribedToMetrics.set(true);
    
    // Start polling for metrics updates
    if (response.pollInterval) {
      startMetricsPolling(response.pollInterval, $api);
    }
    
    console.log('Subscribed to system metrics:', response);
    return response;
  } catch (error) {
    console.error('Failed to subscribe to system metrics:', error);
    throw error;
  }
}

export async function unsubscribeFromMetrics() {
  const $api = get(api);
  const $metricsSubscriptionId = get(metricsSubscriptionId);
  
  if (!$api || !$metricsSubscriptionId) {
    return;
  }
  
  try {
    await $api.unsubscribe($metricsSubscriptionId);
    stopMetricsPolling();
    metricsSubscriptionId.set(null);
    isSubscribedToMetrics.set(false);
    
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
    const $metricsSubscriptionId = get(metricsSubscriptionId);
    const $isConnected = get(isConnected);
    
    if (!$metricsSubscriptionId || !$isConnected) {
      stopMetricsPolling();
      return;
    }
    
    try {
      const response = await apiInstance.pollMetricsUpdates($metricsSubscriptionId);
      if (response.hasData && response.latestMetrics) {
        systemMetrics.set(response.latestMetrics);
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
