import React from 'react';
import { Link } from 'react-router-dom';

const CallHistory: React.FC = () => {
  return (
    <div className="bg-gray-900 text-white min-h-screen p-6">
      <div className="max-w-6xl mx-auto">
        <div className="mb-6">
          <h1 className="text-3xl font-bold text-center">VoIP 通話監控中心</h1>
          <p className="text-gray-400 text-center mt-2">通話紀錄與錄音下載</p>
          
          {/* 返回主頁連結 */}
          <div className="text-center mt-4">
            <Link 
              to="/" 
              className="text-blue-400 hover:text-blue-300 underline text-sm"
            >
              ← 返回主頁
            </Link>
          </div>
        </div>
        
        <div className="bg-gray-800 rounded-lg shadow-lg p-6">
          <h2 className="text-xl font-semibold mb-4">通話紀錄</h2>
          <p className="text-gray-400">載入中...</p>
        </div>
      </div>
    </div>
  );
};

export default CallHistory;
