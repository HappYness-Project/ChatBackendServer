package main

// the code below is from https://medium.com/@parvjn616/building-a-websocket-chat-application-in-go-388fff758575
import (
	"fmt"
	"os"

	"github.com/HappYness-Project/ChatBackendServer/api"
	"github.com/HappYness-Project/ChatBackendServer/configs"
	"github.com/HappYness-Project/ChatBackendServer/dbs"
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
	database, err := dbs.ConnectToDb(connStr)
	if err != nil {
		fmt.Println("Unable to connect to the database." + err.Error())
		return
	}

	server := api.NewApiServer(fmt.Sprintf("%s:%s", env.Host, env.Port), database)
	r := server.Setup()
	if err := server.Run(r); err != nil {
		fmt.Println("Unable to set up the server." + err.Error())
		return
	}
}
