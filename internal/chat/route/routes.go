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
	router.Post("/api/chats/{chatID}/chat-participants", h.AddChatParticipant)
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
	var request CreateChatRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		common.ErrorResponse(w, http.StatusBadRequest, common.ProblemDetails{
			Title:     "Invalid Request Body",
			ErrorCode: "InvalidJSON",
			Detail:    "Unable to decode request body as JSON",
		})
		return
	}

	chatType := domain.ChatTypeGroup
	if request.Type != "" {
		chatType = domain.ChatType(request.Type)
	}
	chat, err := domain.NewChat(chatType, request.UserGroupId, request.ContainerId)
	if err != nil {
		common.ErrorResponse(w, http.StatusBadRequest, common.ProblemDetails{
			Title:     "Invalid Request",
			ErrorCode: "InvalidChatConfiguration",
			Detail:    err.Error(),
		})
		return
	}

	var createdChat *domain.Chat

	if request.ParticipantId != "" {
		participant, err := domain.NewChatParticipant(chat.Id, request.ParticipantId, "admin", "active")
		if err != nil {
			h.logger.Error().Err(err).Msg("Failed to create chat participant")
			common.ErrorResponse(w, http.StatusBadRequest, common.ProblemDetails{
				Title:     "Invalid Request",
				ErrorCode: "InvalidParticipantData",
				Detail:    err.Error(),
			})
			return
		}

		createdChat, err = h.chatRepo.CreateChatWithParticipant(chat, participant)
		if err != nil {
			h.logger.Error().Err(err).Msg("Failed to create chat with participant")
			common.ErrorResponse(w, http.StatusInternalServerError, common.ProblemDetails{
				Title:  "Internal Server Error",
				Detail: "Error occurred while creating chat with participant",
			})
			return
		}
	} else {
		// Create chat only
		createdChat, err = h.chatRepo.CreateChat(chat)
		if err != nil {
			h.logger.Error().Err(err).Msg("Failed to create chat")
			common.ErrorResponse(w, http.StatusInternalServerError, common.ProblemDetails{
				Title:  "Internal Server Error",
				Detail: "Error occurred while creating chat",
			})
			return
		}
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

func (h *Handler) AddChatParticipant(w http.ResponseWriter, r *http.Request) {
	chatID := chi.URLParam(r, "chatID")
	if chatID == "" {
		common.ErrorResponse(w, http.StatusBadRequest, common.ProblemDetails{
			Title:     "Invalid Parameter",
			ErrorCode: "MissingChatID",
			Detail:    "chatID is required",
		})
		return
	}

	var request AddParticipantRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		common.ErrorResponse(w, http.StatusBadRequest, common.ProblemDetails{
			Title:     "Invalid Request Body",
			ErrorCode: "InvalidJSON",
			Detail:    "Unable to decode request body as JSON",
		})
		return
	}

	if request.UserId == "" {
		common.ErrorResponse(w, http.StatusBadRequest, common.ProblemDetails{
			Title:     "Invalid Request",
			ErrorCode: "MissingUserId",
			Detail:    "user_id is required",
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

	// Check if user is already a participant
	isParticipant, err := h.chatRepo.IsUserParticipantInChat(chatID, request.UserId)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to check if user is participant")
		common.ErrorResponse(w, http.StatusInternalServerError, common.ProblemDetails{
			Title:  "Internal Server Error",
			Detail: "Error occurred while checking participant status",
		})
		return
	}

	if isParticipant {
		common.ErrorResponse(w, http.StatusConflict, common.ProblemDetails{
			Title:     "Conflict",
			ErrorCode: "UserAlreadyParticipant",
			Detail:    "User is already a participant in this chat",
		})
		return
	}

	// Create participant using domain constructor (includes validation)
	participant, err := domain.NewChatParticipant(chatID, request.UserId, request.Role, request.Status)
	if err != nil {
		common.ErrorResponse(w, http.StatusBadRequest, common.ProblemDetails{
			Title:     "Invalid Request",
			ErrorCode: "InvalidParticipantData",
			Detail:    err.Error(),
		})
		return
	}

	createdParticipant, err := h.chatRepo.AddParticipantToChat(participant)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to add participant to chat")
		common.ErrorResponse(w, http.StatusInternalServerError, common.ProblemDetails{
			Title:  "Internal Server Error",
			Detail: "Error occurred while adding participant to chat",
		})
		return
	}

	common.WriteJsonWithEncode(w, http.StatusCreated, createdParticipant)
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
