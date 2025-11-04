import { writable, get } from 'svelte/store';
import { newHttpBatchRpcSession } from 'capnweb';

// Configuration
const urlParams = new URLSearchParams(window.location.search);
export const RPC_URL = urlParams.get('rpc_url') || 'http://localhost:8000/rpc';
export const SIMULATED_RTT_MS = Number(urlParams.get('rtt') || 120);
export const SIMULATED_RTT_JITTER_MS = Number(urlParams.get('jitter') || 40);

// Demo state
export const isRunning = writable(false);
export const status = writable('Ready to run demo...');
export const pipelinedResults = writable(null);
export const sequentialResults = writable(null);
export const error = writable(null);

// Utility functions
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const jittered = () => SIMULATED_RTT_MS + (SIMULATED_RTT_JITTER_MS ? Math.random() * SIMULATED_RTT_JITTER_MS : 0);

// Wrap fetch to count RPC POSTs and simulate network latency
const originalFetch = globalThis.fetch;
let fetchCount = 0;

function setupFetchInterceptor() {
  globalThis.fetch = async (input, init) => {
    const method = init?.method || (input instanceof Request ? input.method : 'GET');
    const url = input instanceof Request ? input.url : String(input);
    if (url.startsWith(RPC_URL) && method === 'POST') {
      fetchCount++;
      // Simulate uplink and downlink latency for each RPC POST
      await sleep(jittered());
      const resp = await originalFetch(input, init);
      await sleep(jittered());
      return resp;
    }
    return originalFetch(input, init);
  };
}

// Initialize fetch interceptor
setupFetchInterceptor();

async function runPipelined() {
  fetchCount = 0;
  const t0 = performance.now();

  const api = newHttpBatchRpcSession(RPC_URL);
  const user = api.authenticate('cookie-123');
  const profile = api.getUserProfile(user.id);
  const notifications = api.getNotifications(user.id);

  const [u, p, n] = await Promise.all([user, profile, notifications]);

  const t1 = performance.now();
  console.log(t0, t1, (t1-t0));
  return { u, p, n, ms: t1 - t0, posts: fetchCount };
}

async function runSequential() {
  fetchCount = 0;
  const t0 = performance.now();

  // 1) Authenticate (1 round trip)
  const api1 = newHttpBatchRpcSession(RPC_URL);
  const u = await api1.authenticate('cookie-123');

  // 2) Fetch profile (2nd round trip)
  const api2 = newHttpBatchRpcSession(RPC_URL);
  const p = await api2.getUserProfile(u.id);

  // 3) Fetch notifications (3rd round trip)
  const api3 = newHttpBatchRpcSession(RPC_URL);
  const n = await api3.getNotifications(u.id);

  const t1 = performance.now();
  console.log(t0, t1, (t1-t0));
  return { u, p, n, ms: t1 - t0, posts: fetchCount };
}

export async function runDemo() {
  if (get(isRunning)) return;
  
  isRunning.set(true);
  error.set(null);
  pipelinedResults.set(null);
  sequentialResults.set(null);

  try {
    status.set('Running pipelined demo...');
    const pipelined = await runPipelined();
    pipelinedResults.set(pipelined);

    status.set('Running sequential demo...');
    const sequential = await runSequential();
    sequentialResults.set(sequential);

    status.set('Demo complete!');
  } catch (err) {
    status.set('Error occurred!');
    error.set(err.message);
    console.error(err);
  } finally {
    isRunning.set(false);
  }
}

// Reset function
export function resetDemo() {
  pipelinedResults.set(null);
  sequentialResults.set(null);
  error.set(null);
  status.set('Ready to run demo...');
}
