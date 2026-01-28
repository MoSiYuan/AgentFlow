import { useEffect, useRef, useState, useCallback } from 'react';
import { io, Socket } from 'socket.io-client';

type MessageHandler = (data: any) => void;

interface WebSocketOptions {
  url: string;
  autoConnect?: boolean;
  reconnect?: boolean;
  reconnectInterval?: number;
}

export function useWebSocket(options: WebSocketOptions) {
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const socketRef = useRef<Socket | null>(null);
  const handlersRef = useRef<Map<string, Set<MessageHandler>>>(new Map());

  const connect = useCallback(() => {
    const socket = io(options.url, {
      autoConnect: options.autoConnect ?? true,
      reconnection: options.reconnect ?? true,
      reconnectionDelay: options.reconnectInterval ?? 1000,
    });

    socket.on('connect', () => {
      setIsConnected(true);
      setError(null);
    });

    socket.on('disconnect', () => {
      setIsConnected(false);
    });

    socket.on('error', (err) => {
      setError(err);
    });

    // 消息分发
    socket.onAny((event: string, data: any) => {
      const handlers = handlersRef.current.get(event);
      handlers?.forEach(handler => handler(data));
    });

    socketRef.current = socket;
  }, [options]);

  const disconnect = useCallback(() => {
    socketRef.current?.disconnect();
    socketRef.current = null;
  }, []);

  const on = useCallback((event: string, handler: MessageHandler) => {
    if (!handlersRef.current.has(event)) {
      handlersRef.current.set(event, new Set());
    }
    handlersRef.current.get(event)!.add(handler);

    // 清理函数
    return () => {
      handlersRef.current.get(event)?.delete(handler);
    };
  }, []);

  const emit = useCallback((event: string, data: any) => {
    socketRef.current?.emit(event, data);
  }, []);

  useEffect(() => {
    connect();
    return () => {
      disconnect();
    };
  }, [connect, disconnect]);

  return {
    isConnected,
    error,
    on,
    emit,
    connect,
    disconnect,
  };
}
