package domain

import "fmt"

type ParticipantStatus string

const (
	StatusActive  ParticipantStatus = "active"
	StatusLeft    ParticipantStatus = "left"
	StatusBanned  ParticipantStatus = "banned"
	StatusMuted   ParticipantStatus = "muted"
	StatusPending ParticipantStatus = "pending"
)

func NewParticipantStatus(status string) (ParticipantStatus, error) {
	switch status {
	case string(StatusActive), string(StatusLeft), string(StatusBanned), string(StatusMuted), string(StatusPending):
		return ParticipantStatus(status), nil
	case "":
		return StatusActive, nil // Default to active
	default:
		return "", fmt.Errorf("invalid status: %s. Must be one of: active, left, banned, muted, pending", status)
	}
}

func (s ParticipantStatus) String() string {
	return string(s)
}

func (s ParticipantStatus) IsValid() bool {
	return s == StatusActive || s == StatusLeft || s == StatusBanned || s == StatusMuted || s == StatusPending
}

func (s ParticipantStatus) CanParticipate() bool {
	return s == StatusActive || s == StatusPending
}