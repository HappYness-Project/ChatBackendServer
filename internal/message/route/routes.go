package route

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/HappYness-Project/ChatBackendServer/common"
	domain "github.com/HappYness-Project/ChatBackendServer/internal/message/domain"
	"github.com/HappYness-Project/ChatBackendServer/loggers"

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
	logger      *loggers.AppLogger
	messageRepo msgRepo.MessageRepo
	chatRepo    chatRepo.ChatRepo
}

func NewHandler(logger *loggers.AppLogger, repo msgRepo.MessageRepo, chatRepo chatRepo.ChatRepo) *Handler {
	return &Handler{logger: logger, messageRepo: repo, chatRepo: chatRepo}
}

func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Route("/api", func(r chi.Router) {
		r.Get("/ws/user-groups/{groupID}", h.HandleConnections)
		r.Get("/chats/{chatID}/messages", h.GetMessagesByChatID)
	})
}

func (h *Handler) HandleConnections(w http.ResponseWriter, r *http.Request) {
	groupId, err := strconv.Atoi(chi.URLParam(r, "groupID"))
	if err != nil {
		common.ErrorResponse(w, http.StatusBadRequest, common.ProblemDetails{
			Title:     "Invalid Parameter",
			ErrorCode: "Invalid Group ID",
			Detail:    "The provided groupID is not a valid integer.",
		})
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
		h.logger.Error().Err(err).Msg("Error occurred during getting chat by user group. " + err.Error())
		delete(clients, conn)
		return
	}

	for {
		var msg domain.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			h.logger.Error().Err(err).Msg("Error occurred during reading message. " + err.Error())
			delete(clients, conn)
			return
		}
		msg.ChatID = chat.Id
		msg.MessageType = "text"

		broadcast <- msg
	}
}
func (h *Handler) HandleMessages() {
	for {
		msg := <-broadcast
		if err := h.messageRepo.Create(msg); err != nil {
			// http.Error(w, "Failed to create message", http.StatusInternalServerError)
			h.logger.Error().Err(err).Msg("Unable to create a message")
		}
		fmt.Println(msg.Content)
		//loop through the client list and sending the message to the client
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				h.logger.Error().Err(err).Msg("Unable to write a message")
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func (h *Handler) GetMessagesByChatID(w http.ResponseWriter, r *http.Request) {
	chatID := chi.URLParam(r, "chatID")
	if chatID == "" {
		http.Error(w, "chatID is required", http.StatusBadRequest)
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
