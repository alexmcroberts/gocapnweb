<script>
  export let label;
  export let value;
  export let unit = '%';
  export let progress = 0;
  export let color = '#007acc';
</script>

<div class="metric">
  <span class="metric-label">{label}</span>
  <div class="metric-display">
    <span class="metric-value">{value}{unit}</span>
    <div class="progress-bar">
      <div 
        class="progress-fill" 
        style="width: {progress}%; background: {color}"
      ></div>
    </div>
  </div>
</div>

<style>
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
  
  .metric-display {
    display: flex;
    align-items: center;
  }
  
  .metric-value {
    font-family: 'Monaco', 'Courier New', monospace;
    font-size: 14px;
    font-weight: bold;
  }
  
  .progress-bar {
    width: 100px;
    height: 10px;
    background: #e9ecef;
    border-radius: 5px;
    overflow: hidden;
    margin-left: 10px;
  }
  
  .progress-fill {
    height: 100%;
    transition: width 0.3s ease;
  }
</style>
