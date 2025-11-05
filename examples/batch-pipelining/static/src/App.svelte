<script>
  import { onMount } from 'svelte';
  import ConfigCard from './components/ConfigCard.svelte';
  import ResultsCard from './components/ResultsCard.svelte';
  import SummaryCard from './components/SummaryCard.svelte';
  import { 
    isRunning,
    status,
    pipelinedResults,
    sequentialResults,
    error,
    runDemo,
    resetDemo,
    RPC_URL,
    SIMULATED_RTT_MS,
    SIMULATED_RTT_JITTER_MS
  } from './stores/demoStore.svelte.js';

  // Handle run demo button click
  async function handleRunDemo() {
    await runDemo();
  }

  // Handle reset button click
  function handleReset() {
    resetDemo();
  }

  // Auto-run on mount if enabled
  onMount(() => {
    const urlParams = new URLSearchParams(window.location.search);
    if (urlParams.get('autorun') === 'true') {
      handleRunDemo();
    }
  });
</script>

<main>
  <h1>Cap'n Web RPC - Batch Pipelining Demo (Svelte)</h1>

  <div class="info">
    <h3>What This Demo Shows</h3>
    <p><strong>Pipelining:</strong> Multiple dependent RPC calls are batched into a single HTTP request, dramatically reducing latency.</p>
    <p><strong>Pipeline References:</strong> The result of one call (e.g., <code>user.id</code>) can be used as input to subsequent calls without waiting for the first call to complete.</p>
    <p><strong>Performance:</strong> Compare pipelined vs sequential execution to see the latency benefits.</p>
    <p><strong>Reactive UI:</strong> Built with Svelte for smooth, reactive user experience.</p>
  </div>

  <div class="status" class:error={error()}>
    {#if error()}
      Error: {error()}
      <br><br>
      Make sure the Go server is running on port 8000.
    {:else}
      {status()}
    {/if}
  </div>

  <div class="controls">
    <button 
      onclick={handleRunDemo} 
      disabled={isRunning()}
      class:loading={isRunning()}
    >
      {isRunning() ? 'Running Demo...' : 'Run Demo'}
    </button>
    
    {#if pipelinedResults() || sequentialResults()}
      <button onclick={handleReset} disabled={isRunning()}>
        Reset
      </button>
    {/if}
  </div>

  <div class="results-container">
    <ConfigCard 
      rpcUrl={RPC_URL} 
      rtt={SIMULATED_RTT_MS} 
      jitter={SIMULATED_RTT_JITTER_MS} 
    />

    {#if pipelinedResults()}
      <ResultsCard 
        title="Pipelined (Batched, Single Round Trip)" 
        results={pipelinedResults()} 
      />
    {/if}

    {#if sequentialResults()}
      <ResultsCard 
        title="Sequential (Non-Batched, Multiple Round Trips)" 
        results={sequentialResults()} 
      />
    {/if}

    <SummaryCard 
      pipelined={pipelinedResults()} 
      sequential={sequentialResults()} 
    />
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
  }

  h3 {
    color: #555;
    margin-top: 20px;
  }

  .info {
    background: #e8f4f8;
    border-left: 4px solid #007acc;
    padding: 20px;
    border-radius: 5px;
    margin-bottom: 20px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  }

  .info h3 {
    color: #007acc;
    margin-top: 0;
  }

  .status {
    background: #fff;
    padding: 15px;
    border-radius: 5px;
    margin-bottom: 20px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    font-weight: 500;
  }

  .status.error {
    background: #fee;
    color: #c00;
    border-left: 4px solid #c00;
  }

  .controls {
    margin-bottom: 20px;
  }

  button {
    background: #007acc;
    color: white;
    border: none;
    padding: 10px 20px;
    border-radius: 5px;
    cursor: pointer;
    font-size: 16px;
    margin: 10px 10px 10px 0;
    transition: background-color 0.2s;
  }

  button:hover:not(:disabled) {
    background: #005a99;
  }

  button:disabled {
    background: #ccc;
    cursor: not-allowed;
  }

  button.loading {
    background: #ffa500;
  }

  code {
    background: #f0f0f0;
    padding: 2px 5px;
    border-radius: 3px;
    font-family: 'Monaco', 'Courier New', monospace;
    font-size: 0.9em;
  }

  p {
    margin: 10px 0;
  }

  strong {
    color: #444;
  }

  .results-container {
    display: flex;
    flex-direction: column;
    gap: 0;
  }

  @media (max-width: 768px) {
    :global(body) {
      padding: 10px;
    }
    
    h1 {
      font-size: 1.5rem;
    }
    
    .info {
      padding: 15px;
    }
  }
</style>
