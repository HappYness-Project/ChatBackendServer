package domain

import (
	"fmt"
	"time"
)

type ChatParticipant struct {
	Id       string            `json:"id"`
	ChatId   string            `json:"chat_id"`
	UserId   string            `json:"user_id"`
	JoinedAt time.Time         `json:"joined_at"`
	Role     ParticipantRole   `json:"role"`
	Status   ParticipantStatus `json:"status"`
}

// NewChatParticipant creates a new ChatParticipant with validation
func NewChatParticipant(chatId, userId string, role ParticipantRole, status ParticipantStatus) (*ChatParticipant, error) {
	if chatId == "" {
		return nil, fmt.Errorf("chat_id is required")
	}
	if userId == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	return &ChatParticipant{
		ChatId: chatId,
		UserId: userId,
		Role:   role,
		Status: status,
	}, nil
}

func (cp *ChatParticipant) ChangeRole(newRole string) error {
	role, err := NewParticipantRole(newRole)
	if err != nil {
		return err
	}
	cp.Role = role
	return nil
}

func (cp *ChatParticipant) ChangeStatus(newStatus string) error {
	status, err := NewParticipantStatus(newStatus)
	if err != nil {
		return err
	}
	cp.Status = status
	return nil
}
