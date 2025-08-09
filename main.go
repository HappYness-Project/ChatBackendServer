package main

// the code below is from https://medium.com/@parvjn616/building-a-websocket-chat-application-in-go-388fff758575
import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)

func main() {
	fmt.Println("Run the chat server")

	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", handleConnections)

	go handleMessages()

	fmt.Println("Server started on : 4545")
	err := http.ListenAndServe(":4545", nil)
	if err != nil {
		panic("Error starting server: " + err.Error())
	}

}
func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the Chat Room!")
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

	clients[conn] = true

	for {
		var msg Message
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

func handleMessages() {
	for {
		//got message from channel
		msg := <-broadcast
		fmt.Println(msg.Text)

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

type Message struct {
	Text      string    `json:"text"`
	Sender    string    `json:"sender"`
	Receiver  string    `json:"receiver"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}
