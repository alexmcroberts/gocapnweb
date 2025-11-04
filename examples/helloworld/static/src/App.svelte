<script>
  import { onMount } from 'svelte';
  import { newWebSocketRpcSession } from 'capnweb';
  
  let api = null;
  let isConnected = false;
  let nameInput = 'World';
  let output = 'Ready to connect...';
  let isLoading = false;

  // Initialize the WebSocket connection
  async function initConnection() {
    try {
      // Connect to our Go server's WebSocket endpoint
      api = newWebSocketRpcSession("ws://127.0.0.1:8000/api");
      
      // Test the connection by calling a simple method
      await api.hello("Connection Test");
      isConnected = true;
      
      updateOutput("✅ Connected to Go server successfully!");
      return true;
    } catch (error) {
      updateOutput(`❌ Connection failed: ${error.message}`);
      isConnected = false;
      return false;
    }
  }

  // Function to call the hello method
  async function sayHello() {
    const name = nameInput.trim() || 'World';
    isLoading = true;
    
    try {
      // Ensure we're connected
      if (!isConnected) {
        updateOutput("🔄 Connecting to server...");
        const connected = await initConnection();
        if (!connected) {
          isLoading = false;
          return;
        }
      }
      
      updateOutput(`🔄 Calling hello("${name}")...`);
      
      // Call the server's hello method
      const result = await api.hello(name);
      
      updateOutput(`✅ Server response: "${result}"`);
      
    } catch (error) {
      updateOutput(`❌ Error: ${error.message}`);
      isConnected = false;
    } finally {
      isLoading = false;
    }
  }

  // Function to update the output display
  function updateOutput(message) {
    const timestamp = new Date().toLocaleTimeString();
    output = `[${timestamp}] ${message}\n` + output;
  }

  // Function to clear the output
  function clearOutput() {
    output = 'Output cleared.\n';
  }

  // Handle Enter key in the input field
  function handleKeyPress(event) {
    if (event.key === 'Enter') {
      sayHello();
    }
  }

  // Auto-connect when the component mounts
  onMount(async () => {
    updateOutput("🔄 Initializing connection...");
    await initConnection();
  });
</script>

<main>
  <h1>Hello World - Cap'n Web RPC Svelte Demo</h1>
  
  <div class="info">
    <p><strong>This demo shows:</strong></p>
    <ul>
      <li>WebSocket RPC connection to Go server</li>
      <li>Simple method call with parameters</li>
      <li>Real-time bidirectional communication</li>
      <li>Built with Svelte for reactive UI</li>
    </ul>
  </div>

  <div class="demo-section">
    <h3>WebSocket RPC Demo</h3>
    <p>Enter your name and click "Say Hello" to call the server's hello method:</p>
    
    <div class="input-section">
      <input 
        type="text" 
        bind:value={nameInput}
        on:keypress={handleKeyPress}
        placeholder="Enter your name"
        disabled={isLoading}
      >
      <button 
        on:click={sayHello} 
        disabled={isLoading}
        class:loading={isLoading}
      >
        {isLoading ? 'Calling...' : 'Say Hello'}
      </button>
      <button on:click={clearOutput} disabled={isLoading}>
        Clear Output
      </button>
    </div>
    
    <h4>Output:</h4>
    <div class="output" bind:this={output}>{output}</div>
  </div>
</main>

<style>
  :global(body) {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
    max-width: 800px;
    margin: 0 auto;
    padding: 20px;
    background: #f5f5f5;
  }

  h1 {
    color: #333;
    border-bottom: 2px solid #007acc;
    padding-bottom: 10px;
  }

  .demo-section {
    background: #fff;
    padding: 20px;
    border-radius: 5px;
    margin-bottom: 20px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  }

  .output {
    background: #f0f0f0;
    padding: 15px;
    border-radius: 5px;
    font-family: 'Monaco', 'Courier New', monospace;
    font-size: 14px;
    white-space: pre-wrap;
    min-height: 100px;
    max-height: 300px;
    overflow-y: auto;
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

  button.loading {
    background: #ffa500;
  }

  input {
    padding: 8px;
    border: 1px solid #ddd;
    border-radius: 3px;
    margin: 5px;
    font-size: 14px;
  }

  input:disabled {
    background: #f5f5f5;
    color: #999;
  }

  .input-section {
    margin-bottom: 20px;
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
</style>
