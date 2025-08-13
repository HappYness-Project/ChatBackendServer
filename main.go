package main

// the code below is from https://medium.com/@parvjn616/building-a-websocket-chat-application-in-go-388fff758575
import (
	"fmt"
	"net/http"
	"os"

	"github.com/HappYness-Project/ChatBackendServer/configs"
	"github.com/HappYness-Project/ChatBackendServer/dbs"
	"github.com/HappYness-Project/ChatBackendServer/internal/handler"
)

func main() {
	var current_env = os.Getenv("APP_ENV")
	env := configs.InitConfig(current_env)
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s timezone=UTC connect_timeout=5 ",
		env.DBHost, env.DBPort, env.DBUser, env.DBPwd, env.DBName)
	if current_env == "local" || current_env == "" {
		connStr += "sslmode=disable"
	} else {
		connStr += "sslmode=require"
	}
	fmt.Println("Database connection string: " + connStr)
	db, err := dbs.ConnectToDb(connStr)
	if err != nil {
		// logger.Error().Err(err).Msg("Unable to connect to the database.")
		fmt.Println("Unable to connect to the database." + err.Error())
		return
	}

	// Initialize message repository
	handler.InitMessageRepository(db)
	// Routes
	http.HandleFunc("/", handler.Home)
	http.HandleFunc("/ws", handler.HandleConnections)

	// Message API endpoints
	// http.HandleFunc("/api/messages", handler.CreateMessage)
	// http.HandleFunc("/api/messages/chat", handler.GetMessagesByChatID)
	// http.HandleFunc("/api/messages/user-group", handler.GetMessagesByUserGroup)
	// http.HandleFunc("/api/messages/mark-read", handler.MarkMessageAsRead)
	// http.HandleFunc("/api/messages/delete", handler.DeleteMessage)

	go handler.HandleMessages()
	fmt.Println("Server started on : 4545")
	err = http.ListenAndServe(":4545", nil)
	if err != nil {
		panic("Error starting server: " + err.Error())
	}

}
