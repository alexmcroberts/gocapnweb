// Import the capnweb client library
// In a real deployment, you would install with: npm i capnweb
// For this demo, we'll use a CDN version
import { newWebSocketRpcSession } from 'https://unpkg.com/capnweb@latest/dist/index.js';

let api = null;
let isConnected = false;

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
    const nameInput = document.getElementById('nameInput');
    const name = nameInput.value.trim() || 'World';
    
    try {
        // Ensure we're connected
        if (!isConnected) {
            updateOutput("🔄 Connecting to server...");
            const connected = await initConnection();
            if (!connected) return;
        }
        
        updateOutput(`🔄 Calling hello("${name}")...`);
        
        // Call the server's hello method
        const result = await api.hello(name);
        
        updateOutput(`✅ Server response: "${result}"`);
        
    } catch (error) {
        updateOutput(`❌ Error: ${error.message}`);
        isConnected = false;
    }
}

// Function to update the output display
function updateOutput(message) {
    const output = document.getElementById('output');
    const timestamp = new Date().toLocaleTimeString();
    output.textContent += `[${timestamp}] ${message}\n`;
    output.scrollTop = output.scrollHeight;
}

// Function to clear the output
function clearOutput() {
    document.getElementById('output').textContent = 'Output cleared.\n';
}

// Make functions available globally for HTML onclick handlers
window.sayHello = sayHello;
window.clearOutput = clearOutput;

// Auto-connect when the page loads
document.addEventListener('DOMContentLoaded', async () => {
    updateOutput("🔄 Initializing connection...");
    await initConnection();
});

// Handle Enter key in the input field
document.addEventListener('DOMContentLoaded', () => {
    const nameInput = document.getElementById('nameInput');
    nameInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            sayHello();
        }
    });
});
