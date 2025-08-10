package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// 等待寫入訊息的時間上限。
	writeWait = 10 * time.Second

	// 等待讀取下一則 pong 訊息的時間上限。
	pongWait = 60 * time.Second

	// 向對方發送 ping 訊息的間隔。必須小於 pongWait。
	pingPeriod = (pongWait * 9) / 10

	// 允許從對方讀取的最大訊息大小。
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client 是 WebSocket 伺服器和使用者之間的中間人。
type Client struct {
	hub *Hub

	// WebSocket 連線。
	conn *websocket.Conn

	// 緩衝的傳出訊息 channel。
	send chan []byte

	// 使用者 ID
	userId string
}

// readPump 將訊息從 WebSocket 連線抽送到 hub。
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, rawMessage, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(rawMessage, &msg); err != nil {
			log.Printf("error: unmarshal message: %v", err)
			continue
		}

		messageWithClient := &MessageWithClient{
			msg:    &msg,
			client: c,
		}
		c.hub.route <- messageWithClient
	}
}

// writePump 將訊息從 hub 抽送到 WebSocket 連線。
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// hub 關閉了這個 channel。
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 將佇列中的聊天訊息附加到目前的訊息中。
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWs 處理來自對方的 websocket 請求。
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	// 這裡暫時使用遠端地址作為 userID，後續應改為從請求中獲取真實使用者 ID
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), userId: r.RemoteAddr}
	client.hub.register <- client

	// 允許在 goroutine 中並行處理寫入和讀取
	go client.writePump()
	go client.readPump()
}
