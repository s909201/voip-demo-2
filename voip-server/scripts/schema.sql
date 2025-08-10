-- 使用者資料表
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    voip_id TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 線上使用者資料表
CREATE TABLE IF NOT EXISTS online_users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    ip_address TEXT,
    login_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_active_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users (id)
);

-- 通話紀錄資料表
CREATE TABLE IF NOT EXISTS call_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    caller_id INTEGER NOT NULL,
    receiver_id INTEGER NOT NULL,
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    duration_seconds INTEGER,
    audio_url TEXT,
    status TEXT, -- e.g., 'completed', 'missed', 'failed'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (caller_id) REFERENCES users (id),
    FOREIGN KEY (receiver_id) REFERENCES users (id)
);

-- 為通話紀錄建立索引以優化查詢
CREATE INDEX IF NOT EXISTS idx_caller_id ON call_history (caller_id);
CREATE INDEX IF NOT EXISTS idx_receiver_id ON call_history (receiver_id);
