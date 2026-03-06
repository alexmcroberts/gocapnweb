<script>
  const API_URL = '/api/generate-stream';

  let prompt = $state('Tell me a story');
  let streamedText = $state('');
  let status = $state('idle'); // idle | loading | streaming | done | error
  let errorMessage = $state('');
  let abortController = $state(null);

  async function generate() {
    if (status === 'streaming' || status === 'loading') return;
    streamedText = '';
    errorMessage = '';
    status = 'loading';
    abortController = new AbortController();

    try {
      const response = await fetch(API_URL, {
        method: 'POST',
        headers: { 'Content-Type': 'text/plain' },
        body: prompt || 'Tell me a story',
        signal: abortController.signal,
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      if (!response.body) {
        throw new Error('Response body is not a stream');
      }

      status = 'streaming';
      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let buffer = '';

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        buffer += decoder.decode(value, { stream: true });
        streamedText = buffer;
      }
      status = 'done';
    } catch (err) {
      if (err.name === 'AbortError') {
        status = 'done';
        return;
      }
      status = 'error';
      errorMessage = err.message || String(err);
    } finally {
      abortController = null;
    }
  }

  function stop() {
    if (abortController) {
      abortController.abort();
    }
  }

  function reset() {
    stop();
    streamedText = '';
    errorMessage = '';
    status = 'idle';
  }

  $effect(() => {
    return () => {
      if (abortController) abortController.abort();
    };
  });
</script>

<main>
  <h1>Streamable HTTP – AI typewriter</h1>

  <div class="info">
    <p>Server streams the response with chunked transfer encoding. Each token is sent as soon as it’s ready; the client reads the stream and updates the UI in real time. No WebSocket—plain HTTP.</p>
  </div>

  <div class="controls">
    <label for="prompt">Prompt</label>
    <input
      id="prompt"
      type="text"
      bind:value={prompt}
      placeholder="Tell me a story"
      disabled={status === 'streaming' || status === 'loading'}
    />
    <div class="buttons">
      <button
        onclick={generate}
        disabled={status === 'streaming' || status === 'loading'}
      >
        {status === 'loading' ? 'Starting…' : status === 'streaming' ? 'Streaming…' : 'Generate'}
      </button>
      {#if status === 'streaming'}
        <button onclick={stop} class="stop">Stop</button>
      {/if}
      {#if status !== 'idle' && status !== 'loading'}
        <button onclick={reset} class="secondary">Reset</button>
      {/if}
    </div>
  </div>

  {#if status === 'error'}
    <div class="error">{errorMessage}</div>
  {/if}

  <div class="output-section">
    <h2>Output</h2>
    <div class="output" class:streaming={status === 'streaming'}>
      {#if streamedText}
        {streamedText}
        {#if status === 'streaming'}
          <span class="cursor">|</span>
        {/if}
      {:else if status === 'idle'}
        <span class="placeholder">Click Generate to start streaming.</span>
      {:else if status === 'loading'}
        <span class="placeholder">Connecting…</span>
      {/if}
    </div>
  </div>
</main>

<style>
  :global(body) {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    max-width: 720px;
    margin: 0 auto;
    padding: 24px;
    background: #1a1a1a;
    color: #e0e0e0;
  }

  h1 {
    font-size: 1.5rem;
    margin-bottom: 8px;
    color: #fff;
  }

  h2 {
    font-size: 1rem;
    font-weight: 600;
    color: #aaa;
    margin-bottom: 8px;
  }

  .info {
    background: #252525;
    padding: 12px 16px;
    border-radius: 8px;
    margin-bottom: 20px;
    font-size: 0.9rem;
    line-height: 1.5;
    color: #b0b0b0;
  }

  .controls {
    margin-bottom: 24px;
  }

  .controls label {
    display: block;
    font-size: 0.85rem;
    color: #888;
    margin-bottom: 6px;
  }

  .controls input {
    width: 100%;
    padding: 10px 12px;
    font-size: 1rem;
    border: 1px solid #333;
    border-radius: 6px;
    background: #252525;
    color: #e0e0e0;
    margin-bottom: 12px;
    box-sizing: border-box;
  }

  .controls input:focus {
    outline: none;
    border-color: #0a84ff;
  }

  .buttons {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
  }

  button {
    padding: 10px 18px;
    font-size: 0.95rem;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    background: #0a84ff;
    color: #fff;
  }

  button:hover:not(:disabled) {
    background: #409cff;
  }

  button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  button.stop {
    background: #ff453a;
  }

  button.stop:hover:not(:disabled) {
    background: #ff6961;
  }

  button.secondary {
    background: #3a3a3c;
    color: #e0e0e0;
  }

  button.secondary:hover:not(:disabled) {
    background: #48484a;
  }

  .error {
    background: #3a1a1a;
    color: #ff6961;
    padding: 12px 16px;
    border-radius: 8px;
    margin-bottom: 20px;
  }

  .output-section {
    margin-top: 8px;
  }

  .output {
    background: #252525;
    border: 1px solid #333;
    border-radius: 8px;
    padding: 16px;
    min-height: 120px;
    font-size: 1rem;
    line-height: 1.6;
    white-space: pre-wrap;
    word-break: break-word;
  }

  .output.streaming {
    border-color: #0a84ff;
    box-shadow: 0 0 0 1px rgba(10, 132, 255, 0.2);
  }

  .placeholder {
    color: #666;
  }

  .cursor {
    animation: blink 0.8s step-end infinite;
    color: #0a84ff;
  }

  @keyframes blink {
    50% { opacity: 0; }
  }
</style>
