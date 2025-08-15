package route

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/HappYness-Project/ChatBackendServer/common"
	domain "github.com/HappYness-Project/ChatBackendServer/internal/message/domain"

	chatRepo "github.com/HappYness-Project/ChatBackendServer/internal/chat/repository"
	msgRepo "github.com/HappYness-Project/ChatBackendServer/internal/message/repository"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan domain.Message)
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	messageRepo msgRepo.MessageRepo
	chatRepo    chatRepo.ChatRepo
}

func NewHandler(repo msgRepo.MessageRepo, chatRepo chatRepo.ChatRepo) *Handler {
	return &Handler{messageRepo: repo, chatRepo: chatRepo}
}

func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Route("/api", func(r chi.Router) {
		r.Get("/ws/user-groups/{groupID}", h.HandleConnections)
	})
}

func (h *Handler) HandleConnections(w http.ResponseWriter, r *http.Request) {
	groupId, err := strconv.Atoi(chi.URLParam(r, "groupID"))
	if err != nil {
		// response.ErrorResponse(w, http.StatusBadRequest, *(response.New(constants.InvalidParameter, "Invalid Group ID")))
		return
	}

	// token := r.URL.Query().Get("token")
	// if token == "" {
	// 	http.Error(w, "Missing authentication token", http.StatusUnauthorized)
	// 	return
	// }
	// if !validateJWTToken(token) {
	// 	http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
	// 	return
	// }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return
	}
	defer conn.Close()

	clients[conn] = true
	chat, err := h.chatRepo.GetChatByUserGroupId(groupId)
	if err != nil {
		fmt.Println("Error getting chat by user group ID:", err)
		delete(clients, conn)
		return
	}

	for {
		var msg domain.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println(err)
			delete(clients, conn)
			return
		}
		msg.ChatID = chat.Id
		msg.MessageType = chat.Type
		broadcast <- msg
	}
}

func (h *Handler) HandleMessages() {
	for {
		//got message from channel
		msg := <-broadcast
		if err := h.messageRepo.Create(msg); err != nil {
			// http.Error(w, "Failed to create message", http.StatusInternalServerError)
			fmt.Println(err)
		}
		fmt.Println(msg.Content)
		//loop through the client list and sending the message to the client
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				fmt.Println(err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

// GetMessagesByChatID retrieves messages for a specific chat
func (h *Handler) GetMessagesByChatID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	chatID := r.URL.Query().Get("chat_id")
	if chatID == "" {
		http.Error(w, "chat_id is required", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offsetStr := r.URL.Query().Get("offset")
	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	messages, err := h.messageRepo.GetByChatID(chatID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to retrieve messages", http.StatusInternalServerError)
		return
	}

	common.WriteJsonWithEncode(w, http.StatusOK, map[string]interface{}{
		"messages": messages,
		"count":    len(messages),
	})
}
