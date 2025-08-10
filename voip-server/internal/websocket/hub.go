package websocket

import "encoding/json"

// MessageWithClient 包含一個訊息及其來源客戶端。
type MessageWithClient struct {
	msg    *Message
	client *Client
}

// Hub 維護一組活躍的客戶端，並向客戶端廣播訊息。
type Hub struct {
	// 已註冊的客戶端。
	clients map[string]*Client

	// 用於路由信令訊息。
	route chan *MessageWithClient

	// 從客戶端註冊請求。
	register chan *Client

	// 從客戶端取消註冊請求。
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		route:      make(chan *MessageWithClient),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
	}
}

func (h *Hub) broadcastUserList() {
	var userList []string
	for _, client := range h.clients {
		userList = append(userList, client.userId)
	}

	message, err := json.Marshal(map[string]interface{}{
		"type":  "user_list",
		"users": userList,
	})
	if err != nil {
		// 在實際應用中應處理此錯誤
		return
	}

	for _, client := range h.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client.userId)
		}
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.userId] = client
			h.broadcastUserList()
		case client := <-h.unregister:
			if _, ok := h.clients[client.userId]; ok {
				delete(h.clients, client.userId)
				close(client.send)
				h.broadcastUserList()
			}
		case messageWithClient := <-h.route:
			// 這是子任務 3.3 的核心邏輯
			if targetClient, ok := h.clients[messageWithClient.msg.TargetUserID]; ok {
				// 這是子任務 3.4 的核心邏輯
				rawMessage, err := json.Marshal(messageWithClient.msg)
				if err == nil {
					select {
					case targetClient.send <- rawMessage:
					default:
						close(targetClient.send)
						delete(h.clients, targetClient.userId)
					}
				}
			}
		}
	}
}
