import App from './App.svelte';
import { initializeWebSocket } from './utils/websocket';

// Initialize WebSocket only once
let websocketInitialized = false;

if (!websocketInitialized) {
  initializeWebSocket();
  websocketInitialized = true;
}

const app = new App({
  target: document.body
});

export default app;