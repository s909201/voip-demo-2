package websocket

import "encoding/json"

// Message 定義了客戶端和伺服器之間通訊的結構。
type Message struct {
	Type         string          `json:"type"`
	TargetUserID string          `json:"targetUserId"`
	Payload      json.RawMessage `json:"payload"`
}
