package domain

import "time"

type Chat struct {
	Id          string    `json:"id"`
	Type        string    `json:"type"`
	UserGroupId *int      `json:"usergroup_id,omitempty"`
	ContainerId *string   `json:"container_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
