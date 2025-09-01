# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Local Development
- `make start` - Start the development environment with Docker Compose
- `make down` - Stop and remove Docker containers
- `make logs` - View container logs
- `make rebuild-docker` - Rebuild containers from scratch
- `make watch` - Start with file watching for development
- `make build` - Build the Go application (`go build -v ./...`)
- `make test` - Run unit tests (`go test -v ./... -short`)


### Docker Development Setup

The `dev-env/docker-compose.yml` sets up:
- PostgreSQL database on port 8020
- API server on port 8070
- Auto-initialization with `create_tables.sql`
- Volume mounting for development

## Architecture Overview

This is a real-time chat backend server built in Go with WebSocket support and PostgreSQL database.

### Core Components

**Main Application Structure:**
- `main.go` - Application entry point, sets up database connection and starts API server
- `api/api.go` - HTTP server setup with Chi router, route registration
- Uses environment-based configuration (local/development modes)

**Domain Structure (Clean Architecture):**
- `internal/message/` - Message handling (domain, repository, routes)
- `internal/chat/` - Chat management (domain, repository)
- Each domain has: domain models, repository layer, route handlers

### WebSocket Architecture
- `internal/message/route/socket_manager.go` - WebSocket connection manager
- Real-time message broadcasting to connected clients
- Concurrent client management with mutex protection
- Message queuing with buffered channels

### Database Schema
- **chat** - Chat rooms (private/group/container types)
- **message** - Chat messages with UUID v7 IDs
- **chat_participant** - User membership in chats
- Uses custom UUID v7 generation function for chronological ordering

### Key Technologies
- **Router**: Chi v5 for HTTP routing
- **Database**: PostgreSQL with pgx driver
- **WebSockets**: Gorilla WebSocket
- **Config**: Viper for environment management
- **Logging**: zerolog for structured logging
- **Auth**: JWT tokens for authentication

## Environment Configuration

The application uses different configuration modes:
- **Local/Default**: Uses `dev-env/local.env` file
- **Development**: Reads from environment variables

Database connection automatically switches between SSL modes based on environment.

