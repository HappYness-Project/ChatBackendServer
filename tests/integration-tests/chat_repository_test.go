package integration_tests

import (
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"github.com/HappYness-Project/ChatBackendServer/dbs"
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
		('01987075-16cb-7337-af15-cd28f64c93a4', 'group', NULL, '01987075-16cb-7337-af15-cd28f64c93a4', CURRENT_TIMESTAMP)
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

func TestChatRepository_DatabaseConnection(t *testing.T) {
	repo := repository.NewRepository(testDB)
	require.NotNil(t, repo)

	err := testDB.Ping()
	require.NoError(t, err)
}

func TestChatRepository_TimestampHandling(t *testing.T) {
	setupTestData(t)
	defer cleanupTestData(t)

	repo := repository.NewRepository(testDB)

	t.Run("should return valid timestamp for created_at", func(t *testing.T) {
		chatID := "01987073-0a87-7b32-9439-86868dfe9bd3"

		chat, err := repo.GetChatById(chatID)

		require.NoError(t, err)
		require.NotNil(t, chat)
		assert.False(t, chat.CreatedAt.IsZero())
		assert.True(t, chat.CreatedAt.Before(time.Now().Add(time.Minute)))
	})
}
