<script>
  /**
   * @typedef {Object} Props
   * @property {string} [status] - 'connected', 'connecting', 'disconnected'
   * @property {string} [message]
   */

  /** @type {Props} */
  let { status = 'disconnected', message = '' } = $props();
  
  let statusClass = $derived(`status ${status}`);
</script>

<div class={statusClass}>
  <span>{message}</span>
</div>

<style>
  .status {
    padding: 10px;
    border-radius: 5px;
    margin-bottom: 20px;
    font-weight: 500;
  }
  
  .status.connected {
    background: #d4edda;
    color: #155724;
    border: 1px solid #c3e6cb;
  }
  
  .status.disconnected {
    background: #f8d7da;
    color: #721c24;
    border: 1px solid #f5c6cb;
  }
  
  .status.connecting {
    background: #fff3cd;
    color: #856404;
    border: 1px solid #ffeaa7;
  }
</style>
