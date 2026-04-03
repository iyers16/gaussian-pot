import { useEffect, useRef, useCallback } from 'react';

export type WSEvent = {
  type: string;
  payload: unknown;
};

type Handler = (event: WSEvent) => void;

export function useWebSocket(onEvent: Handler) {
  const wsRef = useRef<WebSocket | null>(null);
  const onEventRef = useRef(onEvent);
  onEventRef.current = onEvent;

  const connect = useCallback(() => {
    const proto = window.location.protocol === 'https:' ? 'wss' : 'ws';
    const url = `${proto}://${window.location.host}/ws`;
    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onmessage = (e) => {
      try {
        const event: WSEvent = JSON.parse(e.data);
        onEventRef.current(event);
      } catch {
        // ignore malformed frames
      }
    };

    ws.onclose = () => {
      // Reconnect after 2s on unexpected close.
      setTimeout(connect, 2000);
    };

    return ws;
  }, []);

  useEffect(() => {
    const ws = connect();
    return () => {
      ws.onclose = null; // suppress reconnect on unmount
      ws.close();
    };
  }, [connect]);
}
