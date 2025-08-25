package repository

import (
	"database/sql"

	"github.com/HappYness-Project/ChatBackendServer/internal/chat/domain"
)

type ChatRepository interface {
	GetChatByUserGroupId(userGroupId int) (*domain.Chat, error)
	GetChatById(chatId string) (*domain.Chat, error)
}

type ChatRepo struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *ChatRepo {
	return &ChatRepo{db: db}
}

func (r *ChatRepo) GetChatById(chatId string) (*domain.Chat, error) {
	rows, err := r.db.Query(`SELECT id, type, usergroup_id, container_id, created_at
							FROM public.chat
							WHERE id = $1`, chatId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	chat := new(domain.Chat)
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

func (r *ChatRepo) GetChatByUserGroupId(userGroupId int) (*domain.Chat, error) {
	rows, err := r.db.Query(`SELECT id, type, usergroup_id, container_id, created_at
							FROM public.chat
							WHERE usergroup_id = $1 and type = 'group'`, userGroupId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chat := new(domain.Chat)
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
func scanRowsIntoChat(rows *sql.Rows) (*domain.Chat, error) {
	chat := new(domain.Chat)
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
