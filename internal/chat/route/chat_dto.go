package route

type AddParticipantRequest struct {
	UserId  string `json:"user_id" validate:"required"`
	GroupId int    `json:"group_id" validate:"required"`
}
type CreateChatRequest struct {
	Type        string  `json:"type"`
	UserGroupId *int    `json:"usergroup_id,omitempty"`
	ContainerId *string `json:"container_id,omitempty"`
	UserId      string  `json:"user_id,omitempty"`
}
