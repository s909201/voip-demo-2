package main

import (
	"log"
	"net/http"
	"strings"
	"voip-server/internal/api"
	"voip-server/internal/config"
	"voip-server/internal/database"
	"voip-server/internal/websocket"
)

func main() {
	// 載入組態
	cfg := config.Load()

	// 連線到資料庫
	db, err := database.ConnectDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("無法連線到資料庫: %v", err)
	}
	defer db.Close()

	// 初始化資料庫結構
	if err := database.InitializeDatabase(db); err != nil {
		log.Fatalf("無法初始化資料庫: %v", err)
	}

	log.Println("資料庫已成功初始化。")

	// 建立 API 處理器
	apiHandlers := api.NewAPIHandlers(db)

	hub := websocket.NewHub()
	go hub.Run()

	// 設定 HTTP 伺服器
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("伺服器運行中"))
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r)
	})

	// API 路由
	http.HandleFunc("/api/upload", apiHandlers.UploadHandler)
	http.HandleFunc("/api/history", apiHandlers.HistoryHandler)
	http.HandleFunc("/api/downloads/", func(w http.ResponseWriter, r *http.Request) {
		// 從 URL 路徑中提取檔案名稱
		path := strings.TrimPrefix(r.URL.Path, "/api/downloads/")
		if path == "" {
			http.Error(w, "檔案名稱不能為空", http.StatusBadRequest)
			return
		}
		apiHandlers.DownloadHandler(w, r)
	})

	log.Printf("正在啟動 HTTPS 伺服器於 https://localhost%s", cfg.ServerPort)
	if err := http.ListenAndServeTLS(cfg.ServerPort, cfg.CertFile, cfg.KeyFile, nil); err != nil {
		log.Fatalf("無法啟動 HTTPS 伺服器: %v", err)
	}
}
