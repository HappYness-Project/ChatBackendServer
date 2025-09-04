package route

type AddParticipantRequest struct {
	UserId string `json:"user_id" validate:"required"`
	Role   string `json:"role,omitempty"`
	Status string `json:"status,omitempty"`
}

type CreateChatRequest struct {
	Type        string  `json:"type"`
	UserGroupId *int    `json:"usergroup_id,omitempty"`
	ContainerId *string `json:"container_id,omitempty"`
}
