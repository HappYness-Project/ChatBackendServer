package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	ID          string    `json:"id"`
	ChatID      string    `json:"chat_id"`
	SenderID    string    `json:"sender_id"`
	Content     string    `json:"content"`
	MessageType string    `json:"message_type"`
	CreatedAt   time.Time `json:"created_at"`
	ReadStatus  bool      `json:"read_status"`
}

func main() {
	// Get username from user
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter your username: ")
	scanner.Scan()
	username := strings.TrimSpace(scanner.Text())

	if username == "" {
		fmt.Println("Username cannot be empty!")
		return
	}

	// Server URL
	serverURL := "ws://localhost:4545/api/ws"

	// Connect to WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		log.Fatal("Failed to connect to WebSocket server:", err)
	}
	defer conn.Close()

	fmt.Printf("Connected to WebSocket server at %s as %s\n", serverURL, username)
	fmt.Println("Type messages to send (or 'quit' to exit):")

	// Channel to handle interrupt signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Channel for incoming messages
	done := make(chan struct{})

	// Goroutine to read messages from server
	go func() {
		defer close(done)
		for {
			var msg Message
			err := conn.ReadJSON(&msg)
			if err != nil {
				log.Println("Read error:", err)
				return
			}
			fmt.Printf("\n[%s] %s: %s\n> ", msg.CreatedAt.Format("15:04:05"), msg.SenderID, msg.Content)
		}
	}()

	// Goroutine to send messages
	go func() {
		inputScanner := bufio.NewScanner(os.Stdin)
		fmt.Print("> ")

		for inputScanner.Scan() {
			text := strings.TrimSpace(inputScanner.Text())

			if text == "quit" {
				fmt.Println("Disconnecting...")
				conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				close(done)
				return
			}

			if text == "" {
				fmt.Print("> ")
				continue
			}

			// Create message
			msg := Message{
				ChatID:      "test-chat-1",
				SenderID:    username,
				Content:     text,
				MessageType: "text",
				CreatedAt:   time.Now(),
			}

			// Send message to server
			err := conn.WriteJSON(msg)
			if err != nil {
				log.Println("Write error:", err)
				return
			}

			fmt.Print("> ")
		}
	}()

	// Wait for interrupt signal or done channel
	select {
	case <-done:
	case <-interrupt:
		fmt.Println("\nInterrupt received, closing connection...")

		// Close connection gracefully
		err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Println("Write close error:", err)
		}

		select {
		case <-done:
		case <-time.After(time.Second):
		}
	}
}
