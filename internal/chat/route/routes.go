package route

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/HappYness-Project/ChatBackendServer/common"
	"github.com/HappYness-Project/ChatBackendServer/internal/chat/domain"
	"github.com/HappYness-Project/ChatBackendServer/internal/chat/repository"
	"github.com/HappYness-Project/ChatBackendServer/loggers"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

type Handler struct {
	logger    *loggers.AppLogger
	chatRepo  repository.ChatRepo
	jwtSecret []byte
}

func NewHandler(logger *loggers.AppLogger, chatRepo repository.ChatRepo, secretKey string) *Handler {
	return &Handler{
		logger:    logger,
		chatRepo:  chatRepo,
		jwtSecret: []byte(secretKey),
	}
}

func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Get("/api/chats/{chatID}", h.GetChatById)
	router.Get("/api/user-groups/{groupID}/chat", h.GetChatByGroupID)
	router.Post("/api/chats", h.CreateChat)
	router.Delete("/api/chats/{chatID}", h.RemoveChat)
	router.Get("/api/chats/{chatID}/chat-participants", h.GetChatParticipants)

}

func (h *Handler) GetChatById(w http.ResponseWriter, r *http.Request) {
	chatID := chi.URLParam(r, "chatID")
	if chatID == "" {
		common.ErrorResponse(w, http.StatusBadRequest, common.ProblemDetails{
			Title:     "Invalid Parameter",
			ErrorCode: "MissingChatID",
			Detail:    "chatID is required",
		})
		return
	}

	chat, err := h.chatRepo.GetChatById(chatID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to retrieve chat by ID")
		common.ErrorResponse(w, http.StatusInternalServerError, common.ProblemDetails{
			Title:  "Internal Server Error",
			Detail: "Error occurred while retrieving chat",
		})
		return
	}

	if chat.Id == "" {
		common.ErrorResponse(w, http.StatusNotFound, common.ProblemDetails{
			Title:     "Not Found",
			ErrorCode: "ChatNotFound",
			Detail:    "Chat not found with the provided ID",
		})
		return
	}

	common.WriteJsonWithEncode(w, http.StatusOK, chat)
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

func (h *Handler) CreateChat(w http.ResponseWriter, r *http.Request) {
	var chat domain.Chat

	if err := json.NewDecoder(r.Body).Decode(&chat); err != nil {
		common.ErrorResponse(w, http.StatusBadRequest, common.ProblemDetails{
			Title:     "Invalid Request Body",
			ErrorCode: "InvalidJSON",
			Detail:    "Unable to decode request body as JSON",
		})
		return
	}
	createdChat, err := h.chatRepo.CreateChat(&chat)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create chat")
		common.ErrorResponse(w, http.StatusInternalServerError, common.ProblemDetails{
			Title:  "Internal Server Error",
			Detail: "Error occurred while creating chat",
		})
		return
	}

	common.WriteJsonWithEncode(w, http.StatusCreated, createdChat)
}

func (h *Handler) RemoveChat(w http.ResponseWriter, r *http.Request) {
	chatID := chi.URLParam(r, "chatID")
	if chatID == "" {
		common.ErrorResponse(w, http.StatusBadRequest, common.ProblemDetails{
			Title:     "Invalid Parameter",
			ErrorCode: "MissingChatID",
			Detail:    "chatID is required",
		})
		return
	}

	chat, err := h.chatRepo.GetChatById(chatID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to retrieve chat by ID")
		common.ErrorResponse(w, http.StatusInternalServerError, common.ProblemDetails{
			Title:  "Internal Server Error",
			Detail: "Error occurred while retrieving chat",
		})
		return
	}

	if chat.Id == "" {
		common.ErrorResponse(w, http.StatusNotFound, common.ProblemDetails{
			Title:     "Not Found",
			ErrorCode: "ChatNotFound",
			Detail:    "Chat not found with the provided ID",
		})
		return
	}

	err = h.chatRepo.DeleteChat(chatID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to delete chat")
		common.ErrorResponse(w, http.StatusInternalServerError, common.ProblemDetails{
			Title:  "Internal Server Error",
			Detail: "Error occurred while deleting chat",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetChatParticipants(w http.ResponseWriter, r *http.Request) {
	chatID := chi.URLParam(r, "chatID")
	if chatID == "" {
		common.ErrorResponse(w, http.StatusBadRequest, common.ProblemDetails{
			Title:     "Invalid Parameter",
			ErrorCode: "MissingChatID",
			Detail:    "chatID is required",
		})
		return
	}

	// Verify chat exists first
	chat, err := h.chatRepo.GetChatById(chatID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to retrieve chat by ID")
		common.ErrorResponse(w, http.StatusInternalServerError, common.ProblemDetails{
			Title:  "Internal Server Error",
			Detail: "Error occurred while retrieving chat",
		})
		return
	}

	if chat.Id == "" {
		common.ErrorResponse(w, http.StatusNotFound, common.ProblemDetails{
			Title:     "Not Found",
			ErrorCode: "ChatNotFound",
			Detail:    "Chat not found with the provided ID",
		})
		return
	}

	participants, err := h.chatRepo.GetChatParticipants(chatID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to retrieve chat participants")
		common.ErrorResponse(w, http.StatusInternalServerError, common.ProblemDetails{
			Title:  "Internal Server Error",
			Detail: "Error occurred while retrieving chat participants",
		})
		return
	}

	common.WriteJsonWithEncode(w, http.StatusOK, map[string]interface{}{
		"participants": participants,
		"count":        len(participants),
	})
}

func (h *Handler) validateJWTToken(tokenString string) bool {
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
