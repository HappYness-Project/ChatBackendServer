package integration_tests

import (
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"github.com/HappYness-Project/ChatBackendServer/dbs"
	"github.com/HappYness-Project/ChatBackendServer/internal/chat/domain"
	"github.com/HappYness-Project/ChatBackendServer/internal/chat/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error
	dsn := "postgres://postgres:postgres@localhost:8020/postgres?sslmode=disable"

	testDB, err = dbs.ConnectToDb(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	code := m.Run()

	if testDB != nil {
		testDB.Close()
	}

	os.Exit(code)
}

func setupTestData(t *testing.T) {
	_, err := testDB.Exec(`
		INSERT INTO public.chat(id, type, usergroup_id, container_id, created_at)
		VALUES
		('01987073-0a87-7b32-9439-86868dfe9bd3', 'group', 100, NULL, CURRENT_TIMESTAMP),
		('01987073-cf13-7621-af36-54ce20056d19', 'group', NULL, NULL, CURRENT_TIMESTAMP),
		('01987075-16cb-7337-af15-cd28f64c93a4', 'group', NULL, NULL, CURRENT_TIMESTAMP)
		ON CONFLICT (id) DO NOTHING
	`)
	require.NoError(t, err)
}

func cleanupTestData(t *testing.T) {
	_, err := testDB.Exec(`
		DELETE FROM public.chat
		WHERE id IN (
			'01987073-0a87-7b32-9439-86868dfe9bd3',
			'01987073-cf13-7621-af36-54ce20056d19',
			'01987075-16cb-7337-af15-cd28f64c93a4'
		)
	`)
	require.NoError(t, err)
}
func TestChatRepository_DatabaseConnection(t *testing.T) {
	repo := repository.NewRepository(testDB)
	require.NotNil(t, repo)

	err := testDB.Ping()
	require.NoError(t, err)
}
func TestChatRepository_GetChatById(t *testing.T) {
	setupTestData(t)
	defer cleanupTestData(t)

	repo := repository.NewRepository(testDB)

	t.Run("should return chat when valid ID provided", func(t *testing.T) {
		chatID := "01987073-0a87-7b32-9439-86868dfe9bd3"

		chat, err := repo.GetChatById(chatID)

		require.NoError(t, err)
		require.NotNil(t, chat)
		assert.Equal(t, chatID, chat.Id)
		assert.Equal(t, "group", chat.Type)
		assert.NotNil(t, chat.UserGroupId)
		assert.Equal(t, 100, *chat.UserGroupId)
		assert.Nil(t, chat.ContainerId)
		assert.False(t, chat.CreatedAt.IsZero())
	})

	t.Run("should return empty chat when non-existent ID provided", func(t *testing.T) {
		nonExistentID := "01987073-0000-0000-0000-000000000000"

		chat, err := repo.GetChatById(nonExistentID)

		require.NoError(t, err)
		require.NotNil(t, chat)
		assert.Empty(t, chat.Id)
	})

}

func TestChatRepository_GetChatByUserGroupId(t *testing.T) {
	setupTestData(t)
	defer cleanupTestData(t)

	repo := repository.NewRepository(testDB)

	t.Run("should return group chat when valid user group ID provided", func(t *testing.T) {
		userGroupID := 100

		chat, err := repo.GetChatByUserGroupId(userGroupID)

		require.NoError(t, err)
		require.NotNil(t, chat)
		assert.Equal(t, "01987073-0a87-7b32-9439-86868dfe9bd3", chat.Id)
		assert.Equal(t, "group", chat.Type)
		assert.NotNil(t, chat.UserGroupId)
		assert.Equal(t, userGroupID, *chat.UserGroupId)
		assert.Nil(t, chat.ContainerId)
		assert.False(t, chat.CreatedAt.IsZero())
	})

	t.Run("should return empty chat when non-existent user group ID provided", func(t *testing.T) {
		nonExistentUserGroupID := 999

		chat, err := repo.GetChatByUserGroupId(nonExistentUserGroupID)

		require.NoError(t, err)
		require.NotNil(t, chat)
		assert.Empty(t, chat.Id)
	})

	t.Run("should use existing test data from schema", func(t *testing.T) {
		userGroupID := 1

		chat, err := repo.GetChatByUserGroupId(userGroupID)

		require.NoError(t, err)
		require.NotNil(t, chat)
		assert.Equal(t, "01987073-0a87-7b32-9439-86868dfe9bd2", chat.Id)
		assert.Equal(t, "group", chat.Type)
		assert.NotNil(t, chat.UserGroupId)
		assert.Equal(t, userGroupID, *chat.UserGroupId)
	})
}

func TestChatRepository_GetChatByGroupID(t *testing.T) {
	setupTestData(t)
	defer cleanupTestData(t)

	repo := repository.NewRepository(testDB)

	t.Run("should return chat when valid group ID provided", func(t *testing.T) {
		groupID := 100

		chat, err := repo.GetChatByGroupID(groupID)

		require.NoError(t, err)
		require.NotNil(t, chat)
		assert.Equal(t, "01987073-0a87-7b32-9439-86868dfe9bd3", chat.Id)
		assert.Equal(t, "group", chat.Type)
		assert.NotNil(t, chat.UserGroupId)
		assert.Equal(t, groupID, *chat.UserGroupId)
		assert.Nil(t, chat.ContainerId)
		assert.False(t, chat.CreatedAt.IsZero())
	})

	t.Run("should return empty chat when non-existent group ID provided", func(t *testing.T) {
		nonExistentGroupID := 999

		chat, err := repo.GetChatByGroupID(nonExistentGroupID)

		require.NoError(t, err)
		require.NotNil(t, chat)
		assert.Empty(t, chat.Id)
	})

	t.Run("should return chat regardless of type (unlike GetChatByUserGroupId)", func(t *testing.T) {
		groupID := 2

		chat, err := repo.GetChatByGroupID(groupID)

		require.NoError(t, err)
		require.NotNil(t, chat)
		assert.Equal(t, "01987073-cf13-7621-af36-54ce20056d18", chat.Id)
		assert.Equal(t, "group", chat.Type)
		assert.NotNil(t, chat.UserGroupId)
		assert.Equal(t, groupID, *chat.UserGroupId)
	})
}

func TestChatRepository_CreateChat(t *testing.T) {
	repo := repository.NewRepository(testDB)

	t.Run("should create group chat successfully", func(t *testing.T) {
		userGroupID := 200
		chat, err := domain.NewChat(domain.ChatTypeGroup, &userGroupID, nil)
		require.NoError(t, err)

		createdChat, err := repo.CreateChat(chat)

		require.NoError(t, err)
		require.NotNil(t, createdChat)
		assert.NotEmpty(t, createdChat.Id)
		assert.Equal(t, domain.ChatTypeGroup, createdChat.Type)
		assert.NotNil(t, createdChat.UserGroupId)
		assert.Equal(t, userGroupID, *createdChat.UserGroupId)
		assert.Nil(t, createdChat.ContainerId)
		assert.False(t, createdChat.CreatedAt.IsZero())
		assert.True(t, createdChat.CreatedAt.After(time.Now().Add(-time.Minute)))

		_, _ = testDB.Exec(`DELETE FROM public.chat WHERE id = $1`, createdChat.Id)
	})
}

func TestChatRepository_DeleteChat(t *testing.T) {
	setupTestData(t)
	defer cleanupTestData(t)

	repo := repository.NewRepository(testDB)

	t.Run("should delete existing chat successfully", func(t *testing.T) {
		chatID := "01987073-0a87-7b32-9439-86868dfe9bd3"

		// Verify chat exists before deletion
		chat, err := repo.GetChatById(chatID)
		require.NoError(t, err)
		require.NotNil(t, chat)
		assert.Equal(t, chatID, chat.Id)

		// Delete the chat
		err = repo.DeleteChat(chatID)
		require.NoError(t, err)

		// Verify chat no longer exists
		deletedChat, err := repo.GetChatById(chatID)
		require.NoError(t, err)
		require.NotNil(t, deletedChat)
		assert.Empty(t, deletedChat.Id) // Should return empty chat when not found
	})

	t.Run("should handle deletion of non-existent chat gracefully", func(t *testing.T) {
		nonExistentID := "01987073-0000-0000-0000-000000000000"

		err := repo.DeleteChat(nonExistentID)
		require.NoError(t, err) // Should not error even if chat doesn't exist
	})

	t.Run("should delete chat created during test", func(t *testing.T) {
		// Create a chat first
		userGroupID := 300
		chat := &domain.Chat{
			Type:        "group",
			UserGroupId: &userGroupID,
		}

		createdChat, err := repo.CreateChat(chat)
		require.NoError(t, err)
		require.NotNil(t, createdChat)

		// Verify it was created
		foundChat, err := repo.GetChatById(createdChat.Id)
		require.NoError(t, err)
		assert.Equal(t, createdChat.Id, foundChat.Id)

		// Delete it
		err = repo.DeleteChat(createdChat.Id)
		require.NoError(t, err)

		// Verify it's deleted
		deletedChat, err := repo.GetChatById(createdChat.Id)
		require.NoError(t, err)
		assert.Empty(t, deletedChat.Id)
	})

	t.Run("should not affect other chats when deleting one", func(t *testing.T) {
		// Create two chats
		chat1 := &domain.Chat{Type: "private"}
		chat2 := &domain.Chat{Type: "private"}

		createdChat1, err1 := repo.CreateChat(chat1)
		createdChat2, err2 := repo.CreateChat(chat2)
		require.NoError(t, err1)
		require.NoError(t, err2)

		// Delete only the first one
		err := repo.DeleteChat(createdChat1.Id)
		require.NoError(t, err)

		// Verify first is deleted
		deletedChat, err := repo.GetChatById(createdChat1.Id)
		require.NoError(t, err)
		assert.Empty(t, deletedChat.Id)

		// Verify second still exists
		existingChat, err := repo.GetChatById(createdChat2.Id)
		require.NoError(t, err)
		assert.Equal(t, createdChat2.Id, existingChat.Id)

		// Cleanup
		_, _ = testDB.Exec(`DELETE FROM public.chat WHERE id = $1`, createdChat2.Id)
	})
}

func TestChatRepository_GetChatParticipants(t *testing.T) {
	repo := repository.NewRepository(testDB)

	t.Run("should return participants for existing chat", func(t *testing.T) {
		chatID := "01987073-0a87-7b32-9439-86868dfe9bd2"

		participants, err := repo.GetChatParticipants(chatID)

		require.NoError(t, err)
		require.NotNil(t, participants)
		assert.Greater(t, len(participants), 0)

		// Verify participant structure
		for _, p := range participants {
			assert.NotEmpty(t, p.Id)
			assert.Equal(t, chatID, p.ChatId)
			assert.NotEmpty(t, p.UserId)
			assert.Contains(t, []string{"admin", "member"}, p.Role)
			assert.Contains(t, []string{"active", "left", "banned", "muted", "pending"}, p.Status)
			assert.False(t, p.JoinedAt.IsZero())
		}
	})

	t.Run("should return empty slice for non-existent chat", func(t *testing.T) {
		nonExistentID := "01987073-0000-0000-0000-000000000000"

		participants, err := repo.GetChatParticipants(nonExistentID)

		require.NoError(t, err)
		require.NotNil(t, participants)
		assert.Equal(t, 0, len(participants))
	})

	t.Run("should return participants ordered by joined_at ASC", func(t *testing.T) {
		// Create a chat with multiple participants at different times
		userGroupID := 400
		chat := &domain.Chat{
			Type:        "group",
			UserGroupId: &userGroupID,
		}

		createdChat, err := repo.CreateChat(chat)
		require.NoError(t, err)

		// Insert participants with different joined_at times
		_, err = testDB.Exec(`INSERT INTO public.chat_participant(chat_id, user_id, joined_at, role, status) VALUES
			($1, '01959b38-b3f9-7ec5-8ac8-e353bfe08a2d', '2024-01-02 00:00:00', 'admin', 'active'),
			($1, '01959b39-febd-770d-9e1b-e5ee392fce54', '2024-01-01 00:00:00', 'member', 'active')`,
			createdChat.Id)
		require.NoError(t, err)

		participants, err := repo.GetChatParticipants(createdChat.Id)
		require.NoError(t, err)
		assert.Len(t, participants, 2)

		// Should be ordered by joined_at ASC (earliest first)
		assert.True(t, participants[0].JoinedAt.Before(participants[1].JoinedAt))

		// Cleanup
		_, _ = testDB.Exec(`DELETE FROM public.chat_participant WHERE chat_id = $1`, createdChat.Id)
		_, _ = testDB.Exec(`DELETE FROM public.chat WHERE id = $1`, createdChat.Id)
	})
}
