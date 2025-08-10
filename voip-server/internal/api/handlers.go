package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CallHistory 代表通話紀錄結構
type CallHistory struct {
	ID        int       `json:"id"`
	CallID    string    `json:"call_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	AudioURL  string    `json:"audio_url"`
}

// APIHandlers 包含所有 API 處理器和依賴項
type APIHandlers struct {
	DB *sql.DB
}

// NewAPIHandlers 建立新的 API 處理器實例
func NewAPIHandlers(db *sql.DB) *APIHandlers {
	return &APIHandlers{
		DB: db,
	}
}

// UploadHandler 處理錄音檔案上傳
func (h *APIHandlers) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "方法不被允許", http.StatusMethodNotAllowed)
		return
	}

	// 解析 multipart form data，限制最大記憶體使用量為 32MB
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "無法解析表單資料: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 獲取 call_id
	callID := r.FormValue("callId")
	if callID == "" {
		http.Error(w, "缺少 callId 參數", http.StatusBadRequest)
		return
	}

	// 獲取上傳的檔案
	file, header, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, "無法獲取音訊檔案: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 產生檔案名稱（使用 call_id + .wav）
	filename := callID + ".wav"

	// 確保上傳目錄存在
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		http.Error(w, "無法建立上傳目錄: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 建立完整的檔案路徑
	filePath := filepath.Join(uploadDir, filename)

	// 建立目標檔案
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "無法建立檔案: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// 複製檔案內容
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "無法儲存檔案: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 開始資料庫交易處理「先到先得」邏輯
	tx, err := h.DB.Begin()
	if err != nil {
		// 刪除已儲存的檔案
		os.Remove(filePath)
		http.Error(w, "無法開始資料庫交易: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback() // 如果沒有明確提交，則回滾

	// 檢查該 call_id 的 audio_url 是否已存在
	var existingAudioURL sql.NullString
	err = tx.QueryRow("SELECT audio_url FROM call_history WHERE call_id = ?", callID).Scan(&existingAudioURL)

	if err != nil && err != sql.ErrNoRows {
		// 資料庫查詢錯誤
		os.Remove(filePath)
		http.Error(w, "資料庫查詢錯誤: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err == sql.ErrNoRows {
		// 該 call_id 不存在，需要先建立紀錄
		// 這種情況下我們先建立一個基本的通話紀錄
		_, err = tx.Exec("INSERT INTO call_history (call_id, audio_url) VALUES (?, ?)",
			callID, "/api/downloads/"+filename)
		if err != nil {
			os.Remove(filePath)
			http.Error(w, "無法建立通話紀錄: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else if existingAudioURL.Valid && existingAudioURL.String != "" {
		// audio_url 已存在，表示已有錄音檔案，忽略此次上傳
		os.Remove(filePath) // 刪除剛才儲存的檔案
		tx.Commit()         // 提交交易（雖然沒有變更）

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		response := fmt.Sprintf(`{"message": "該通話已有錄音檔案", "existing_file": "%s"}`, existingAudioURL.String)
		w.Write([]byte(response))
		return
	} else {
		// audio_url 為空，更新為新的檔案路徑
		_, err = tx.Exec("UPDATE call_history SET audio_url = ? WHERE call_id = ?",
			"/api/downloads/"+filename, callID)
		if err != nil {
			os.Remove(filePath)
			http.Error(w, "無法更新通話紀錄: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// 提交交易
	err = tx.Commit()
	if err != nil {
		os.Remove(filePath)
		http.Error(w, "無法提交資料庫交易: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 回傳成功訊息
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := fmt.Sprintf(`{"message": "檔案上傳成功", "filename": "%s", "size": %d}`, filename, header.Size)
	w.Write([]byte(response))
}

// HistoryHandler 處理通話紀錄查詢
func (h *APIHandlers) HistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "方法不被允許", http.StatusMethodNotAllowed)
		return
	}

	// 查詢所有通話紀錄，按開始時間降序排列
	rows, err := h.DB.Query("SELECT id, call_id, start_time, end_time, audio_url FROM call_history ORDER BY start_time DESC")
	if err != nil {
		http.Error(w, "資料庫查詢錯誤: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// 建立切片來儲存查詢結果
	var histories []CallHistory

	// 掃描查詢結果
	for rows.Next() {
		var history CallHistory
		var startTime, endTime sql.NullTime
		var audioURL sql.NullString

		err := rows.Scan(&history.ID, &history.CallID, &startTime, &endTime, &audioURL)
		if err != nil {
			http.Error(w, "掃描資料錯誤: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 處理可能為 NULL 的時間欄位
		if startTime.Valid {
			history.StartTime = startTime.Time
		}
		if endTime.Valid {
			history.EndTime = endTime.Time
		}
		if audioURL.Valid {
			history.AudioURL = audioURL.String
		}

		histories = append(histories, history)
	}

	// 檢查是否有掃描錯誤
	if err = rows.Err(); err != nil {
		http.Error(w, "掃描結果錯誤: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 將結果序列化為 JSON
	jsonData, err := json.Marshal(histories)
	if err != nil {
		http.Error(w, "JSON 序列化錯誤: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 設定回應標頭並回傳 JSON 資料
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// DownloadHandler 處理錄音檔案下載
func (h *APIHandlers) DownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "方法不被允許", http.StatusMethodNotAllowed)
		return
	}

	// 從 URL 路徑中解析檔案名稱
	// 假設路由格式為 /api/downloads/{filename}
	path := r.URL.Path
	if !strings.HasPrefix(path, "/api/downloads/") {
		http.Error(w, "無效的下載路徑", http.StatusBadRequest)
		return
	}

	// 提取檔案名稱
	filename := strings.TrimPrefix(path, "/api/downloads/")
	if filename == "" {
		http.Error(w, "缺少檔案名稱", http.StatusBadRequest)
		return
	}

	// 防止路徑遍歷攻擊，清理檔案名稱
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		http.Error(w, "無效的檔案名稱", http.StatusBadRequest)
		return
	}

	// 建立完整的檔案路徑
	uploadDir := "./uploads"
	filePath := filepath.Join(uploadDir, filename)

	// 檢查檔案是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "檔案不存在", http.StatusNotFound)
		return
	}

	// 設定適當的 Content-Type 標頭
	w.Header().Set("Content-Type", "audio/wav")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// 使用 http.ServeFile 提供檔案下載
	http.ServeFile(w, r, filePath)
}
