<script>
  export let title;
  export let results;
  export let showDetails = false;
  
  function toggleDetails() {
    showDetails = !showDetails;
  }
</script>

<div class="results">
  <h3>{title}</h3>
  <p><strong>HTTP POSTs:</strong> {results.posts}</p>
  <p><strong>Time:</strong> {results.ms.toFixed(2)} ms</p>
  
  <details bind:open={showDetails}>
    <summary on:click|preventDefault={toggleDetails}>Data received</summary>
    {#if showDetails}
      <p><strong>Authenticated user:</strong> <code>{JSON.stringify(results.u)}</code></p>
      <p><strong>Profile:</strong> <code>{JSON.stringify(results.p)}</code></p>
      <p><strong>Notifications:</strong> <code>{JSON.stringify(results.n)}</code></p>
    {/if}
  </details>
</div>

<style>
  .results {
    background: #fff;
    padding: 20px;
    border-radius: 5px;
    margin-bottom: 20px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  }
  
  .results h3 {
    color: #007acc;
    margin-top: 0;
  }
  
  code {
    background: #f0f0f0;
    padding: 2px 5px;
    border-radius: 3px;
    font-family: 'Monaco', 'Courier New', monospace;
    font-size: 0.9em;
  }
  
  details {
    margin-top: 10px;
  }
  
  summary {
    cursor: pointer;
    color: #007acc;
    font-weight: 500;
  }
  
  summary:hover {
    text-decoration: underline;
  }
  
  p {
    margin: 10px 0;
  }
  
  strong {
    color: #444;
  }
</style>
