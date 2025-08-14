import { writable } from 'svelte/store';

export const socket = writable(null);

export function connectSocket(url) {
  const ws = new WebSocket(url);

  ws.onopen = () => console.log('[WS] Connected');
  ws.onclose = () => {
    console.log('[WS] Disconnected, retrying in 1s...');
    setTimeout(() => connectSocket(url), 1000); // auto-reconnect
  };
  ws.onerror = (err) => console.error('[WS] Error', err);

  socket.set(ws);
}
