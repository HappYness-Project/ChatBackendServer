package handler

import (
	"fmt"
	"net/http"

	"github.com/HappYness-Project/ChatBackendServer/internal/message"
	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan message.Message)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return
	}
	defer conn.Close()

	clients[conn] = true

	for {
		var msg message.Message
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
