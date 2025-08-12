# WebSocket Chat Client

A simple WebSocket client for connecting to the ChatBackendServer.

## Usage

1. Make sure the main server is running on port 4545:
   ```bash
   cd ..
   go run main.go
   ```

2. Run the client:
   ```bash
   go run main.go
   ```

3. Start typing messages to send to the chat server. Type 'quit' to exit.

## Features

- Connect to WebSocket server at `ws://localhost:4545/ws`
- Send and receive real-time messages
- Graceful disconnect with 'quit' command
- Display incoming messages with timestamps

## Message Format

The client sends messages in the following JSON format:
```json
{
  "chat_id": "test-chat-1",
  "sender_id": "client-user", 
  "content": "your message",
  "message_type": "text",
  "created_at": "2023-01-01T00:00:00Z"
}
```