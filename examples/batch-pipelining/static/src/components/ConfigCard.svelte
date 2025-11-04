<script>
  export let rpcUrl;
  export let rtt;
  export let jitter;
</script>

<div class="config">
  <h3>Configuration</h3>
  <p><strong>RPC URL:</strong> {rpcUrl}</p>
  <p><strong>Simulated RTT (each direction):</strong> ~{rtt}ms ±{jitter}ms</p>
  <p><strong>Server:</strong> Go implementation of Cap'n Web RPC</p>
</div>

<style>
  .config {
    background: #fff;
    padding: 20px;
    border-radius: 5px;
    margin-bottom: 20px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  }
  
  .config h3 {
    color: #555;
    margin-top: 0;
  }
  
  p {
    margin: 10px 0;
  }
  
  strong {
    color: #444;
  }
</style>
