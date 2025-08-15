package route

import (
	"fmt"
	"net/http"
	"sync"

	domain "github.com/HappYness-Project/ChatBackendServer/internal/message/domain"
	"github.com/HappYness-Project/ChatBackendServer/loggers"
	"github.com/gorilla/websocket"
)

type WebSocketManager struct {
	clients   map[*websocket.Conn]bool
	broadcast chan domain.Message
	upgrader  websocket.Upgrader
	mutex     sync.RWMutex
}

func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan domain.Message, 256),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (wsm *WebSocketManager) AddClient(conn *websocket.Conn) {
	wsm.mutex.Lock()
	defer wsm.mutex.Unlock()
	wsm.clients[conn] = true
}

func (wsm *WebSocketManager) RemoveClient(conn *websocket.Conn) {
	wsm.mutex.Lock()
	defer wsm.mutex.Unlock()
	delete(wsm.clients, conn)
	conn.Close()
}

func (wsm *WebSocketManager) BroadcastMessage(msg domain.Message) {
	select {
	case wsm.broadcast <- msg:
	default:
		fmt.Println("Broadcast channel full, dropping message")
	}
}

func (wsm *WebSocketManager) SendToClients(msg domain.Message, logger *loggers.AppLogger) {
	wsm.mutex.RLock()
	clientsCopy := make([]*websocket.Conn, 0, len(wsm.clients))
	for client := range wsm.clients {
		clientsCopy = append(clientsCopy, client)
	}
	wsm.mutex.RUnlock()

	for _, client := range clientsCopy {
		err := client.WriteJSON(msg)
		if err != nil {
			logger.Error().Err(err).Msg("Unable to write a message")
			wsm.RemoveClient(client)
		}
	}
}
