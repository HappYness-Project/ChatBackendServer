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
func NewChatParticipant(chatId, userId string, roleStr, statusStr string) (*ChatParticipant, error) {
	if chatId == "" {
		return nil, fmt.Errorf("chat_id is required")
	}
	if userId == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	role, err := NewParticipantRole(roleStr)
	if err != nil {
		return nil, err
	}

	status, err := NewParticipantStatus(statusStr)
	if err != nil {
		return nil, err
	}

	return &ChatParticipant{
		ChatId: chatId,
		UserId: userId,
		Role:   role,
		Status: status,
	}, nil
}

// ChangeRole changes the participant's role with validation
func (cp *ChatParticipant) ChangeRole(newRole string) error {
	role, err := NewParticipantRole(newRole)
	if err != nil {
		return err
	}
	cp.Role = role
	return nil
}

// ChangeStatus changes the participant's status with validation
func (cp *ChatParticipant) ChangeStatus(newStatus string) error {
	status, err := NewParticipantStatus(newStatus)
	if err != nil {
		return err
	}
	cp.Status = status
	return nil
}

// CanParticipate returns whether the participant can actively participate
func (cp *ChatParticipant) CanParticipate() bool {
	return cp.Status.CanParticipate()
}

// IsAdmin returns whether the participant is an admin
func (cp *ChatParticipant) IsAdmin() bool {
	return cp.Role == RoleAdmin
}