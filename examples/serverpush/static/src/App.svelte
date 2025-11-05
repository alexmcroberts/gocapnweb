<script>
  import ConnectionStatus from './components/ConnectionStatus.svelte';
  import MetricCard from './components/MetricCard.svelte';
  import { 
    getIsConnected,
    getConnectionStatus,
    getConnectionMessage,
    getSystemMetrics,
    getMetricsSubscriptionId,
    getCanConnect,
    getCanSubscribe,
    getCanUnsubscribe,
    connectToServer,
    subscribeToMetrics,
    unsubscribeFromMetrics,
    cleanup
  } from './stores/appStore.svelte.js';

  let isLoading = $state(false);

  // Handle connect button click
  async function handleConnect() {
    isLoading = true;
    try {
      await connectToServer();
    } finally {
      isLoading = false;
    }
  }

  // Handle subscribe button click
  async function handleSubscribe() {
    isLoading = true;
    try {
      await subscribeToMetrics();
    } catch (error) {
      alert(`Failed to subscribe: ${error.message}`);
    } finally {
      isLoading = false;
    }
  }

  // Handle unsubscribe button click
  async function handleUnsubscribe() {
    isLoading = true;
    try {
      await unsubscribeFromMetrics();
    } catch (error) {
      console.error('Failed to unsubscribe:', error);
    } finally {
      isLoading = false;
    }
  }

  // Format timestamp
  function formatTimestamp(timestamp) {
    if (!timestamp) return 'Never';
    return new Date(timestamp * 1000).toLocaleTimeString();
  }

  // Auto-connect when the component mounts and cleanup on destroy
  $effect(() => {
    console.log('🚀 Server Push Demo loaded');
    handleConnect();
    
    // Return cleanup function
    return () => {
      cleanup();
    };
  });
</script>

<main>
  <h1>💻 System Metrics Demo - Real-time Monitoring</h1>
  
  <div class="info">
    <p><strong>This demo showcases:</strong></p>
    <ul>
      <li>Live system metrics with automatic updates</li>
      <li>Subscription-based data feeds</li>
      <li>Real-time data visualization</li>
      <li>WebSocket Server Push technology</li>
      <li>Built with Svelte for reactive UI</li>
      <li>Real CPU, Disk, and Network monitoring via gopsutil</li>
    </ul>
  </div>

  <ConnectionStatus status={getConnectionStatus()} message={getConnectionMessage()} />

  <div class="demo-section">
    <h3>Connection Controls</h3>
    <button 
      onclick={handleConnect} 
      disabled={!getCanConnect() || isLoading}
      class:loading={isLoading}
    >
      {isLoading ? 'Connecting...' : 'Connect'}
    </button>
    <button 
      onclick={handleSubscribe} 
      disabled={!getCanSubscribe() || isLoading}
      class:loading={isLoading}
    >
      {isLoading ? 'Subscribing...' : 'Subscribe to System Metrics'}
    </button>
    <button 
      onclick={handleUnsubscribe} 
      disabled={!getCanUnsubscribe() || isLoading}
      class="danger"
    >
      Unsubscribe
    </button>
  </div>

  <div class="demo-section">
    <h3>💻 System Metrics</h3>
    <div class="subscription-status">
      {#if getMetricsSubscriptionId()}
        ✅ Subscribed (ID: {getMetricsSubscriptionId()})
      {:else}
        Not subscribed
      {/if}
    </div>
    
    <div class="chart-container">
      <MetricCard 
        label="CPU Usage" 
        value={getSystemMetrics().cpuPercent.toFixed(1)} 
        unit="%" 
        progress={getSystemMetrics().cpuPercent}
        color="linear-gradient(90deg, #28a745, #ffc107, #dc3545)"
      />
      
      <MetricCard 
        label="Disk Usage" 
        value={getSystemMetrics().diskUsage.toFixed(1)} 
        unit="%" 
        progress={getSystemMetrics().diskUsage}
        color="linear-gradient(90deg, #28a745, #ffc107, #dc3545)"
      />
      
      <MetricCard 
        label="Network I/O" 
        value={getSystemMetrics().networkIO.toFixed(1)} 
        unit=" (scaled)" 
        progress={getSystemMetrics().networkIO}
        color="linear-gradient(90deg, #17a2b8, #007bff, #6f42c1)"
      />
      
      <div class="metric">
        <span class="metric-label">Last Updated</span>
        <span class="timestamp">{formatTimestamp(getSystemMetrics().timestamp)}</span>
      </div>
    </div>
  </div>
</main>

<style>
  :global(body) {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
    max-width: 1200px;
    margin: 0 auto;
    padding: 20px;
    background: #f5f5f5;
  }

  h1 {
    color: #333;
    border-bottom: 2px solid #007acc;
    padding-bottom: 10px;
    margin-bottom: 30px;
  }

  .demo-section {
    background: #fff;
    padding: 20px;
    border-radius: 8px;
    margin-bottom: 20px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  }

  .chart-container {
    background: #f8f9fa;
    border-radius: 5px;
    padding: 15px;
    display: flex;
    flex-direction: column;
  }

  button {
    background: #007acc;
    color: white;
    border: none;
    padding: 10px 20px;
    border-radius: 5px;
    cursor: pointer;
    font-size: 16px;
    margin: 10px 5px 10px 0;
    transition: background-color 0.2s;
  }

  button:hover:not(:disabled) {
    background: #005a99;
  }

  button:disabled {
    background: #ccc;
    cursor: not-allowed;
  }

  button.danger {
    background: #dc3545;
  }

  button.danger:hover:not(:disabled) {
    background: #c82333;
  }

  button.loading {
    background: #ffa500;
  }

  .info {
    background: #e8f4f8;
    padding: 15px;
    border-radius: 5px;
    border-left: 4px solid #007acc;
    margin-bottom: 20px;
  }

  .info ul {
    margin: 10px 0 0 0;
    padding-left: 20px;
  }

  .info li {
    margin-bottom: 5px;
  }

  .subscription-status {
    margin-bottom: 15px;
    font-weight: 500;
  }

  .timestamp {
    font-size: 12px;
    color: #666;
    font-family: 'Monaco', 'Courier New', monospace;
  }

  .metric {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 10px 0;
    border-bottom: 1px solid #eee;
  }

  .metric:last-child {
    border-bottom: none;
  }

  .metric-label {
    font-weight: 500;
  }

  @media (max-width: 768px) {
    :global(body) {
      padding: 10px;
    }
    
    h1 {
      font-size: 1.5rem;
    }
    
    .demo-section {
      padding: 15px;
    }
  }
</style>
