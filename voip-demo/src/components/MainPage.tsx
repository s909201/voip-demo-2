import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { useWebSocket } from '../hooks/useWebSocket';
import { useWebRTC } from '../hooks/useWebRTC';

const MainPage: React.FC = () => {
  const [username, setUsername] = useState('');
  const { connectionStatus, onlineUsers, connect, disconnect } = useWebSocket();
  const { callState, uploadStatus, resetUploadStatus } = useWebRTC();
  const isConnected = connectionStatus === 'connected';

  const handleConnect = () => {
    if (username.trim()) {
      connect(username);
    }
  };

  const handleDisconnect = () => {
    disconnect();
  };

  const statusMessages: { [key: string]: string } = {
    connecting: '連線中...',
    connected: `已連線為 ${username}`,
    disconnected: '未連線',
    error: '連線錯誤',
  };

  const uploadMessages: { [key: string]: string } = {
    idle: '',
    uploading: '錄音上傳中...',
    success: '錄音上傳成功！',
    error: '錄音上傳失敗，請稍後重試',
  };

  const getUploadStatusColor = (status: string) => {
    switch (status) {
      case 'uploading':
        return 'text-yellow-400';
      case 'success':
        return 'text-green-400';
      case 'error':
        return 'text-red-400';
      default:
        return 'text-gray-300';
    }
  };

  return (
    <div className="bg-gray-900 text-white min-h-screen flex items-center justify-center font-sans">
      <main className="bg-gray-800 rounded-lg shadow-lg p-6 w-full max-w-sm">
        <div className="mb-6 text-center">
          <h1 className="text-2xl font-bold">VoIP Demo</h1>
          <p className="text-gray-400 text-sm mt-1">簡易 WebRTC 語音通話</p>
        </div>

        {/* 導航連結 */}
        <div className="mb-4 text-center">
          <Link 
            to="/history" 
            className="text-blue-400 hover:text-blue-300 underline text-sm"
          >
            查看通話紀錄
          </Link>
        </div>

        {/* 使用者資訊區 */}
        <div className="mb-4">
          <label htmlFor="username" className="block text-sm font-medium text-gray-300 mb-1">你的名字</label>
          <div className="flex">
            <input
              type="text"
              id="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="flex-grow bg-gray-700 border border-gray-600 rounded-l-md p-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="輸入你的名字..."
              disabled={isConnected}
            />
            <button
              onClick={isConnected ? handleDisconnect : handleConnect}
              className={`px-4 py-2 rounded-r-md font-semibold transition-colors ${
                isConnected
                  ? 'bg-red-600 hover:bg-red-700'
                  : 'bg-green-600 hover:bg-green-700'
              }`}
              disabled={!username.trim() && !isConnected}
            >
              {isConnected ? '離線' : '連線'}
            </button>
          </div>
        </div>

        {/* 通話控制區 */}
        <div className="mb-4">
          <label htmlFor="contact" className="block text-sm font-medium text-gray-300 mb-1">選擇聯絡人</label>
          <select id="contact" className="w-full bg-gray-700 border border-gray-600 rounded-md p-2 focus:outline-none focus:ring-2 focus:ring-blue-500" disabled={!isConnected || onlineUsers.length === 0}>
            <option>-- {onlineUsers.length > 0 ? '選擇一個線上使用者' : '沒有線上使用者'} --</option>
            {onlineUsers.map(user => (
              <option key={user} value={user}>{user}</option>
            ))}
          </select>
          <div className="flex justify-around mt-4">
            <button className="bg-blue-600 hover:bg-blue-700 text-white font-bold py-2 px-6 rounded-full transition-colors disabled:bg-gray-500" disabled={!isConnected}>
              通話
            </button>
            <button className="bg-red-600 hover:bg-red-700 text-white font-bold py-2 px-6 rounded-full transition-colors disabled:bg-gray-500" disabled={!isConnected}>
              掛斷
            </button>
          </div>
        </div>

        {/* 狀態顯示區 */}
        <div className="mb-4 text-center p-2 bg-gray-700 rounded-md">
          <p className="text-sm text-gray-300">{statusMessages[connectionStatus]}</p>
          {uploadStatus !== 'idle' && (
            <p className={`text-sm mt-1 ${getUploadStatusColor(uploadStatus)}`}>
              {uploadMessages[uploadStatus]}
              {uploadStatus === 'error' && (
                <button
                  onClick={resetUploadStatus}
                  className="ml-2 text-xs underline hover:no-underline"
                >
                  清除
                </button>
              )}
            </p>
          )}
        </div>

        {/* 下載按鈕區 */}
        <div>
          <button className="w-full bg-purple-600 hover:bg-purple-700 text-white font-bold py-2 px-4 rounded-md transition-colors disabled:bg-gray-500" disabled={!isConnected}>
            下載錄音
          </button>
        </div>
      </main>
    </div>
  );
};

export default MainPage;
