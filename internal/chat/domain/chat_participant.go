package domain

import "time"

type ChatParticipant struct {
	Id       string    `json:"id"`
	ChatId   string    `json:"chat_id"`
	UserId   string    `json:"user_id"`
	JoinedAt time.Time `json:"joined_at"`
	Role     string    `json:"role"`
	Status   string    `json:"status"`
}