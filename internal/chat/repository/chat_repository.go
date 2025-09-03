package repository

import (
	"database/sql"
	"time"

	"github.com/HappYness-Project/ChatBackendServer/internal/chat/domain"
	"github.com/google/uuid"
)

type ChatRepository interface {
	GetChatByUserGroupId(userGroupId int) (*domain.Chat, error)
	GetChatById(chatId string) (*domain.Chat, error)
	GetChatByGroupID(groupID int) (*domain.Chat, error)
	CreateChat(chat *domain.Chat) (*domain.Chat, error)
	DeleteChat(chatId string) error
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

func (r *ChatRepo) GetChatByGroupID(groupID int) (*domain.Chat, error) {
	rows, err := r.db.Query(`SELECT id, type, usergroup_id, container_id, created_at
							FROM public.chat
							WHERE usergroup_id = $1`, groupID)
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

func (r *ChatRepo) CreateChat(chat *domain.Chat) (*domain.Chat, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	
	chat.Id = id.String()
	chat.CreatedAt = time.Now().UTC()
	
	_, err = r.db.Exec(`INSERT INTO public.chat (id, type, usergroup_id, container_id, created_at) 
						VALUES ($1, $2, $3, $4, $5)`,
		chat.Id, chat.Type, chat.UserGroupId, chat.ContainerId, chat.CreatedAt)
	if err != nil {
		return nil, err
	}
	
	return chat, nil
}

func (r *ChatRepo) DeleteChat(chatId string) error {
	_, err := r.db.Exec(`DELETE FROM public.chat WHERE id = $1`, chatId)
	return err
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
