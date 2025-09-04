package domain

import "fmt"

type ParticipantRole string

const (
	RoleAdmin  ParticipantRole = "admin"
	RoleMember ParticipantRole = "member"
)

func NewParticipantRole(role string) (ParticipantRole, error) {
	switch role {
	case string(RoleAdmin), string(RoleMember):
		return ParticipantRole(role), nil
	case "":
		return RoleMember, nil // Default to member
	default:
		return "", fmt.Errorf("invalid role: %s. Must be 'admin' or 'member'", role)
	}
}

func (r ParticipantRole) String() string {
	return string(r)
}

func (r ParticipantRole) IsValid() bool {
	return r == RoleAdmin || r == RoleMember
}