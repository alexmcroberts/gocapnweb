<script>
  export let pipelined;
  export let sequential;
  
  $: speedup = sequential && pipelined ? (sequential.ms / pipelined.ms) : 0;
</script>

<div class="summary">
  <h3>Summary</h3>
  {#if pipelined && sequential}
    <p><strong>Pipelined:</strong> {pipelined.posts} POST, {pipelined.ms.toFixed(2)} ms</p>
    <p><strong>Sequential:</strong> {sequential.posts} POSTs, {sequential.ms.toFixed(2)} ms</p>
    <p><strong>Speedup:</strong> {speedup.toFixed(2)}x faster with pipelining!</p>
    <p><strong>Key Insight:</strong> Pipelining allows dependent calls to be batched together, eliminating multiple round trips and dramatically reducing latency in high-latency networks.</p>
  {:else}
    <p>Run the demo to see performance comparison...</p>
  {/if}
</div>

<style>
  .summary {
    background: #e8f4f8;
    border-left: 4px solid #007acc;
    padding: 20px;
    border-radius: 5px;
    margin-bottom: 20px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  }
  
  .summary h3 {
    color: #007acc;
    margin-top: 0;
  }
  
  p {
    margin: 10px 0;
  }
  
  strong {
    color: #444;
  }
</style>
