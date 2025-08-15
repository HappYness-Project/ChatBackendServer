package repository

import (
	"database/sql"

	entity "github.com/HappYness-Project/ChatBackendServer/internal/entity"
)

type MessageRepository interface {
	Create(message entity.Message) error
	GetChatByUserGroupId(userGroupId int) (*entity.Chat, error)
	GetByChatID(chatID string, limit, offset int) ([]entity.Message, error)
	GetByUserGroup(userIDs []string, limit, offset int) ([]entity.Message, error)
}

type MessageRepo struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *MessageRepo {
	return &MessageRepo{db: db}
}

func (r *MessageRepo) Create(message entity.Message) error {
	query := `
		INSERT INTO message (chat_id, sender_id, content, message_type, created_at)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(query, message.ChatID, message.SenderID, message.Content, message.MessageType, message.CreatedAt)

	return err
}

// Reminder : This is just a temporary method we are going to use it to get chat info by user group id
// In the future, we will use the identity service to get the chat info
func (r *MessageRepo) GetChatByUserGroupId(userGroupId int) (*entity.Chat, error) {
	rows, err := r.db.Query(`SELECT id, type, usergroup_id, container_id, created_at
							FROM public.chat
							WHERE usergroup_id = $1 and type = 'group'`, userGroupId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chat := new(entity.Chat)
	for rows.Next() {
		chat, err = scanRowsIntoChat(rows)
		if err != nil {
			return nil, err
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return chat, nil
}
func scanRowsIntoChat(rows *sql.Rows) (*entity.Chat, error) {
	chat := new(entity.Chat)
	err := rows.Scan(
		&chat.Id,
		&chat.Type,
		&chat.UserGroupId,
		&chat.ContainerId,
		&chat.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return chat, nil
}
func (r *MessageRepo) GetByChatID(chatID string, limit, offset int) ([]entity.Message, error) {
	query := `
		SELECT id, chat_id, sender_id, content, message_type, created_at, read_status
		FROM messages
		WHERE chat_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, chatID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []entity.Message
	for rows.Next() {
		var msg entity.Message
		err := rows.Scan(&msg.ID, &msg.ChatID, &msg.SenderID, &msg.Content,
			&msg.MessageType, &msg.CreatedAt, &msg.ReadStatus)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}
func (r *MessageRepo) GetByUserGroup(userIDs []string, limit, offset int) ([]entity.Message, error) {
	if len(userIDs) == 0 {
		return []entity.Message{}, nil
	}

	query := `
		SELECT DISTINCT m.id, m.chat_id, m.sender_id, m.content, m.message_type, m.created_at, m.read_status
		FROM messages m
		INNER JOIN chat_participants cp ON m.chat_id = cp.chat_id
		WHERE cp.user_id = ANY($1)
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, userIDs, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []entity.Message
	for rows.Next() {
		var msg entity.Message
		err := rows.Scan(&msg.ID, &msg.ChatID, &msg.SenderID, &msg.Content,
			&msg.MessageType, &msg.CreatedAt, &msg.ReadStatus)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}
