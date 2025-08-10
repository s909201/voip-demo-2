import { useState, useRef, useCallback } from 'react';

type ConnectionStatus = 'disconnected' | 'connecting' | 'connected' | 'error';

// 定義從伺服器收到的訊息格式
interface ServerMessage {
  type: string;
  users?: string[];
  // 其他可能的 payload 欄位
}

export const useWebSocket = () => {
  const [connectionStatus, setConnectionStatus] = useState<ConnectionStatus>('disconnected');
  const [onlineUsers, setOnlineUsers] = useState<string[]>([]);
  const ws = useRef<WebSocket | null>(null);

  const connect = useCallback((username: string) => {
    if (ws.current || connectionStatus === 'connecting') {
      return;
    }

    setConnectionStatus('connecting');
    
    // 注意：在實際部署中，應使用 wss:// 和正確的主機名稱
    const wsUrl = `wss://${window.location.host}/ws`;
    ws.current = new WebSocket(wsUrl);

    ws.current.onopen = () => {
      setConnectionStatus('connected');
      // 登入訊息，後端需要相應的邏輯來處理
      const loginMessage = { type: 'login', payload: { userId: username } };
      ws.current?.send(JSON.stringify(loginMessage));
    };

    ws.current.onmessage = (event) => {
      const message: ServerMessage = JSON.parse(event.data);
      if (message.type === 'user_list' && message.users) {
        // 過濾掉自己
        setOnlineUsers(message.users.filter(user => user !== username));
      }
      // 在此處可以處理其他類型的訊息，例如 offer, answer, candidate
    };

    ws.current.onerror = () => {
      setConnectionStatus('error');
    };

    ws.current.onclose = () => {
      setConnectionStatus('disconnected');
      ws.current = null;
    };

  }, [connectionStatus]);

  const disconnect = useCallback(() => {
    ws.current?.close();
  }, []);

  return {
    connectionStatus,
    onlineUsers,
    connect,
    disconnect,
    send: ws.current?.send.bind(ws.current),
  };
};
