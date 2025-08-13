package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	entity "github.com/HappYness-Project/ChatBackendServer/internal/entity"
	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan entity.Message)
var messageRepo *Repository
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func InitMessageRepository(db *sql.DB) {
	messageRepo = NewRepository(db)
}

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return
	}
	defer conn.Close()

	clients[conn] = true
	// TODO : API request to the Task API to get the detail information
	newFunction()

	//TODO needs
	for {
		var msg entity.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println(err)
			delete(clients, conn)
			return
		}

		broadcast <- msg
	}

	// for {
	// 	var msg Message
	// 	err := conn.ReadJSON(&msg)
	// 	if err != nil {
	// 		return
	// 	}

	// 	if msg.Type == "new_client" {
	// 		//add new user
	// 	} else if msg.Type == "chat" {
	// 		//chat broadcast
	// 	} else if msg.Type == "session_end" {
	// 		//remove client
	// 	} else {
	// 		//log.Println("Unknown message:", msg.Type)
	// 	}
	// }
}

func newFunction() {
	externalAPIURL := "https://example.com/api/user-groups/1/chats" // Replace with actual Task API URL
	req, err := http.NewRequest("GET", externalAPIURL, nil)
	if err != nil {
		fmt.Println("Error creating request to external API:", err)
	} else {
		// Optionally add headers, authentication, etc.
		// req.Header.Set("Authorization", "Bearer <token>")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error making request to external API:", err)
		} else {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				var apiResponse map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
					fmt.Println("Error decoding external API response:", err)
				} else {
					fmt.Println("External API response:", apiResponse)
					// You can use apiResponse as needed
				}
			} else {
				fmt.Printf("External API returned status: %d\n", resp.StatusCode)
			}
		}
	}
}

func HandleMessages() {
	for {
		//got message from channel
		msg := <-broadcast
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
func Home(w http.ResponseWriter, r *http.Request) {
	var payload = struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Version string `json:"version"`
	}{
		Status:  "active",
		Message: "Message Service server",
		Version: "1.0.0",
	}
	WriteJsonWithEncode(w, http.StatusOK, payload)
}

func CreateMessage(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return
	}
	defer conn.Close()

	var req entity.CreateMessageDto
	err = conn.ReadJSON(&req)
	if err != nil {
		fmt.Println("ReadJSON error:", err)
		return
	}

	if req.ChatID == "" || req.SenderID == "" || req.Content == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}
	msg := entity.Message{
		ChatID:      req.ChatID,
		SenderID:    req.SenderID,
		Content:     req.Content,
		MessageType: req.MessageType,
	}

	if err := messageRepo.Create(msg); err != nil {
		http.Error(w, "Failed to create message", http.StatusInternalServerError)
		return
	}

	// Broadcast to WebSocket clients
	select {
	case broadcast <- msg: // I changed it to object(non-pointer... but should it be?)
	default:
	}

	WriteJsonWithEncode(w, http.StatusCreated, msg)
}

// GetMessagesByChatID retrieves messages for a specific chat
func GetMessagesByChatID(w http.ResponseWriter, r *http.Request) {
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

	messages, err := messageRepo.GetByChatID(chatID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to retrieve messages", http.StatusInternalServerError)
		return
	}

	WriteJsonWithEncode(w, http.StatusOK, map[string]interface{}{
		"messages": messages,
		"count":    len(messages),
	})
}

// GetMessagesByUserGroup retrieves messages for a group of users
func GetMessagesByUserGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDsStr := r.URL.Query().Get("user_ids")
	if userIDsStr == "" {
		http.Error(w, "user_ids is required", http.StatusBadRequest)
		return
	}

	userIDs := strings.Split(userIDsStr, ",")
	for i, id := range userIDs {
		userIDs[i] = strings.TrimSpace(id)
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

	messages, err := messageRepo.GetByUserGroup(userIDs, limit, offset)
	if err != nil {
		http.Error(w, "Failed to retrieve messages", http.StatusInternalServerError)
		return
	}

	WriteJsonWithEncode(w, http.StatusOK, map[string]interface{}{
		"messages": messages,
		"count":    len(messages),
	})
}

// MarkMessageAsRead marks a message as read
func MarkMessageAsRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		MessageID string `json:"message_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.MessageID == "" {
		http.Error(w, "message_id is required", http.StatusBadRequest)
		return
	}

	if err := messageRepo.UpdateReadStatus(req.MessageID, true); err != nil {
		http.Error(w, "Failed to mark message as read", http.StatusInternalServerError)
		return
	}

	WriteJsonWithEncode(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Message marked as read",
	})
}

// DeleteMessage deletes a message
func DeleteMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	messageID := r.URL.Query().Get("message_id")
	if messageID == "" {
		http.Error(w, "message_id is required", http.StatusBadRequest)
		return
	}

	if err := messageRepo.Delete(messageID); err != nil {
		http.Error(w, "Failed to delete message", http.StatusInternalServerError)
		return
	}

	WriteJsonWithEncode(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Message deleted",
	})
}
