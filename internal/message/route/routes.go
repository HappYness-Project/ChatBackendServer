package route

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/HappYness-Project/ChatBackendServer/common"
	domain "github.com/HappYness-Project/ChatBackendServer/internal/message/domain"
	"github.com/HappYness-Project/ChatBackendServer/loggers"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	chatRepo "github.com/HappYness-Project/ChatBackendServer/internal/chat/repository"
	msgRepo "github.com/HappYness-Project/ChatBackendServer/internal/message/repository"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	logger      *loggers.AppLogger
	messageRepo msgRepo.MessageRepo
	chatRepo    chatRepo.ChatRepo
	wsManager   *WebSocketManager
	jwtSecret   []byte
}

func NewHandler(logger *loggers.AppLogger, repo msgRepo.MessageRepo, chatRepo chatRepo.ChatRepo, secretKey string) *Handler {
	wsManager := NewWebSocketManager(logger)
	handler := &Handler{
		logger:      logger,
		messageRepo: repo,
		chatRepo:    chatRepo,
		wsManager:   wsManager,
		jwtSecret:   []byte(secretKey),
	}
	go handler.HandleMessages()
	return handler
}

func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Route("/api", func(r chi.Router) {
		r.Get("/user-groups/{groupID}/ws", h.HandleConnectionsByGroupId)
		r.Get("/chats/{chatID}/ws", h.HandleConnectionsByChatID)
		r.Get("/chats/{chatID}/messages", h.GetMessagesByChatID)
		r.Get("/user-groups/{groupID}/messages", h.GetMessagesByGroupID)
		r.Get("/user-groups/{groupID}/chat", h.GetChatByGroupID)
	})
}

func (h *Handler) HandleConnectionsByChatID(w http.ResponseWriter, r *http.Request) {
	if !h.authenticateRequest(w, r) {
		common.ErrorResponse(w, http.StatusUnauthorized, common.ProblemDetails{
			Title:     "Unauthorized",
			ErrorCode: "AuthenticationFailure",
			Detail:    "Invalid authentication token",
		})
		return
	}

	chatID := chi.URLParam(r, "chatID")
	if chatID == "" {
		h.logger.Error().Msg("chatID is required")
		http.Error(w, "chatID is required", http.StatusBadRequest)
		return
	}

	conn, err := h.wsManager.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error().Err(err).Msg(err.Error())
		return
	}
	defer h.wsManager.RemoveClient(conn)

	h.wsManager.AddClient(conn)
	chat, err := h.chatRepo.GetChatById(chatID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Error occurred during getting chat by user group. " + err.Error())
		return
	}

	for {
		var msg domain.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			h.logger.Error().Err(err).Msg("Error occurred during reading message. " + err.Error())
			return
		}
		msg.ChatID = chat.Id
		msg.CreatedAt = time.Now().UTC()
		msg.MessageType = "text"

		h.wsManager.BroadcastMessage(msg)
	}
}

func (h *Handler) HandleConnectionsByGroupId(w http.ResponseWriter, r *http.Request) {
	if !h.authenticateRequest(w, r) {
		common.ErrorResponse(w, http.StatusUnauthorized, common.ProblemDetails{
			Title:     "Unauthorized",
			ErrorCode: "AuthenticationFailure",
			Detail:    "Invalid authentication token",
		})
		return
	}

	groupId, err := strconv.Atoi(chi.URLParam(r, "groupID"))
	if err != nil {
		h.logger.Error().Msg("JWT token validation failure")
		common.ErrorResponse(w, http.StatusBadRequest, common.ProblemDetails{
			Title:     "Invalid Parameter",
			ErrorCode: "Invalid Group ID",
			Detail:    "The provided groupID is not a valid integer.",
		})
		return
	}
	conn, err := h.wsManager.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error().Err(err).Msg(err.Error())
		return
	}
	defer h.wsManager.RemoveClient(conn)

	h.wsManager.AddClient(conn)
	chat, err := h.chatRepo.GetChatByUserGroupId(groupId)
	if err != nil {
		h.logger.Error().Err(err).Msg("Error occurred during getting chat by user group. " + err.Error())
		return
	}

	for {
		var msg domain.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			h.logger.Error().Err(err).Msg("Error occurred during reading message. " + err.Error())
		}
		msg.ChatID = chat.Id
		msg.CreatedAt = time.Now().UTC()
		msg.MessageType = "text"

		h.wsManager.BroadcastMessage(msg)
	}
}
func (h *Handler) HandleMessages() {
	for {
		msg := <-h.wsManager.broadcast
		id, _ := uuid.NewV7()
		msg.ID = id.String()
		if err := h.messageRepo.Create(msg); err != nil {
			h.logger.Error().Err(err).Msg("Unable to create a message")
			continue
		}
		fmt.Printf("Broadcasting message to %d clients\n", len(h.wsManager.clients))
		fmt.Printf("[ChatID:%s]|[SenderID:%s]|Message: %s\n", msg.ChatID, msg.SenderID, msg.Content)
		fmt.Println("-------------------------------------------------------------")
		h.wsManager.SendToClients(msg, h.logger)
	}
}

func (h *Handler) GetMessagesByChatID(w http.ResponseWriter, r *http.Request) {
	chatID := chi.URLParam(r, "chatID")
	if chatID == "" {
		http.Error(w, "chatID is required", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 120
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
		h.logger.Error().Err(err).Msg("Failed to retrieve messages by chatID")
		common.ErrorResponse(w, http.StatusInternalServerError, common.ProblemDetails{
			Title:  "Internal Server Error",
			Detail: "Error occurred during getting chat ID",
		})
		return
	}

	common.WriteJsonWithEncode(w, http.StatusOK, map[string]interface{}{
		"messages": messages,
		"count":    len(messages),
	})
}

func (h *Handler) GetMessagesByGroupID(w http.ResponseWriter, r *http.Request) {
	groupIDStr := chi.URLParam(r, "groupID")
	if groupIDStr == "" {
		http.Error(w, "groupID is required", http.StatusBadRequest)
		return
	}

	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		common.ErrorResponse(w, http.StatusBadRequest, common.ProblemDetails{
			Title:     "Invalid Parameter",
			ErrorCode: "Invalid Group ID",
			Detail:    "The provided groupID is not a valid integer.",
		})
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

	chat, err := h.chatRepo.GetChatByUserGroupId(groupID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to retrieve chat by groupID")
		common.ErrorResponse(w, http.StatusInternalServerError, common.ProblemDetails{
			Title:  "Internal Server Error",
			Detail: "Error occurred during getting chat by group ID",
		})
		return
	}

	messages, err := h.messageRepo.GetByChatID(chat.Id, limit, offset)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to retrieve messages by groupID")
		common.ErrorResponse(w, http.StatusInternalServerError, common.ProblemDetails{
			Title:  "Internal Server Error",
			Detail: "Error occurred during getting messages by group ID",
		})
		return
	}

	common.WriteJsonWithEncode(w, http.StatusOK, map[string]interface{}{
		"messages": messages,
		"count":    len(messages),
	})
}

func (h *Handler) GetChatByGroupID(w http.ResponseWriter, r *http.Request) {
	groupIDStr := chi.URLParam(r, "groupID")
	if groupIDStr == "" {
		http.Error(w, "groupID is required", http.StatusBadRequest)
		return
	}

	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		common.ErrorResponse(w, http.StatusBadRequest, common.ProblemDetails{
			Title:     "Invalid Parameter",
			ErrorCode: "Invalid Group ID",
			Detail:    "The provided groupID is not a valid integer.",
		})
		return
	}

	chat, err := h.chatRepo.GetChatByGroupID(groupID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to retrieve chat by groupID")
		common.ErrorResponse(w, http.StatusInternalServerError, common.ProblemDetails{
			Title:  "Internal Server Error",
			Detail: "Error occurred during getting chat by group ID",
		})
		return
	}

	common.WriteJsonWithEncode(w, http.StatusOK, chat)
}
func (h *Handler) authenticateRequest(_ http.ResponseWriter, r *http.Request) bool {
	token := r.URL.Query().Get("token")
	if token == "" {
		h.logger.Error().Msg("Missing jwt token")
		return false
	}

	if !h.validateJWTToken(token) {
		h.logger.Error().Msg("Invalid jwt token")
		return false
	}
	return true
}

func (h *Handler) validateJWTToken(tokenString string) bool {
	// TODO: Implement real JWT validation
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS512 {
			return nil, fmt.Errorf("unexpected signing method: %v, expected HS512", token.Header["alg"])
		}
		return h.jwtSecret, nil
	})

	if err != nil {
		h.logger.Error().Err(err).Msg("JWT parsing error")
		return false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		h.logger.Info().Interface("claims", claims).Msg("JWT token validated successfully")
		return true
	}

	h.logger.Error().Msg("Invalid JWT token claims")
	return false
}
