package handler

import (
	"database/sql"

	entity "github.com/HappYness-Project/ChatBackendServer/internal/entity"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(message entity.Message) error {
	query := `
		INSERT INTO messages (chat_id, sender_id, content, message_type, created_at)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(query, message.ChatID, message.SenderID, message.Content, message.MessageType, message.CreatedAt)

	return err
}

func (r *Repository) GetByChatID(chatID string, limit, offset int) ([]entity.Message, error) {
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

func (r *Repository) GetByUserGroup(userIDs []string, limit, offset int) ([]entity.Message, error) {
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

func (r *Repository) GetByID(messageID string) (*entity.Message, error) {
	query := `
		SELECT id, chat_id, sender_id, content, message_type, created_at, read_status
		FROM messages
		WHERE id = $1
	`

	var msg entity.Message
	err := r.db.QueryRow(query, messageID).Scan(
		&msg.ID, &msg.ChatID, &msg.SenderID, &msg.Content,
		&msg.MessageType, &msg.CreatedAt, &msg.ReadStatus)

	if err != nil {
		return nil, err
	}

	return &msg, nil
}

func (r *Repository) UpdateReadStatus(messageID string, readStatus bool) error {
	query := `UPDATE messages SET read_status = $1 WHERE id = $2`
	_, err := r.db.Exec(query, readStatus, messageID)
	return err
}

func (r *Repository) Delete(messageID string) error {
	query := `DELETE FROM messages WHERE id = $1`
	_, err := r.db.Exec(query, messageID)
	return err
}
