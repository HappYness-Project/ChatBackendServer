package greetings

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	fmt.Println("Hello world")
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return
	}
	defer conn.Close()
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			return
		}

		if msg.Type == "new_client" {
			//add new user
		} else if msg.Type == "chat" {
			//chat broadcast
		} else if msg.Type == "session_end" {
			//remove client
		} else {
			//log.Println("Unknown message:", msg.Type)
		}
	}
}

func sendMessage(c *websocket.Conn, clientID string) {

}

func addNewUser() {

}

type Message struct {
	Text      string    `json:"text"`
	Sender    string    `json:"sender"`
	Receiver  string    `json:"receiver"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}
