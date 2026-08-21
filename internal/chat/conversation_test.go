package chat

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cogniforge/internal/database"
	"cogniforge/internal/model"
)

func setupConversationService(t *testing.T) *ConversationService {
	t.Helper()
	db := database.InitTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.ChatConversation{}))
	return NewConversationService(db)
}

func TestTitleFromMessages(t *testing.T) {
	assert.Equal(t, "", titleFromMessages(nil))
	assert.Equal(t, "hello", titleFromMessages([]model.ConversationMessage{
		{Role: "assistant", Content: "ignore"},
		{Role: "user", Content: "  hello  "},
	}))

	long := stringsRepeat("你", 50)
	got := titleFromMessages([]model.ConversationMessage{{Role: "user", Content: long}})
	assert.Equal(t, stringsRepeat("你", 40)+"…", got)

}

func stringsRepeat(s string, n int) string {
	out := make([]rune, 0, n)
	r := []rune(s)[0]
	for i := 0; i < n; i++ {
		out = append(out, r)
	}
	return string(out)
}

func TestConversationCRUDAndIsolation(t *testing.T) {
	svc := setupConversationService(t)

	created, err := svc.Create("user-a", &CreateConversationRequest{
		AgentID: "agent-1",
		Model:   "deepseek-chat",
		Messages: []model.ConversationMessage{
			{ID: "m1", Role: "user", Content: "今天天气怎么样"},
			{ID: "m2", Role: "assistant", Content: "晴"},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "今天天气怎么样", created.Title)
	assert.Equal(t, "user-a", created.UserID)
	require.Len(t, created.Messages, 2)

	listed, err := svc.List("user-a")
	require.NoError(t, err)
	require.Len(t, listed, 1)
	assert.Equal(t, created.ID, listed[0].ID)

	otherList, err := svc.List("user-b")
	require.NoError(t, err)
	assert.Empty(t, otherList)

	_, err = svc.Get("user-b", created.ID)
	require.ErrorIs(t, err, errConversationNotFound)

	got, err := svc.Get("user-a", created.ID)
	require.NoError(t, err)
	require.Len(t, got.Messages, 2)
	assert.Equal(t, "晴", got.Messages[1].Content)

	msgs := []model.ConversationMessage{
		{ID: "m1", Role: "user", Content: "今天天气怎么样"},
		{ID: "m2", Role: "assistant", Content: "晴"},
		{ID: "m3", Role: "user", Content: "带伞吗"},
	}
	updated, err := svc.Update("user-a", created.ID, &UpdateConversationRequest{
		Messages: &msgs,
		Model:    strPtr("deepseek-chat"),
	})
	require.NoError(t, err)
	require.Len(t, updated.Messages, 3)
	assert.Equal(t, "今天天气怎么样", updated.Title)

	require.ErrorIs(t, svc.Delete("user-b", created.ID), errConversationNotFound)
	require.NoError(t, svc.Delete("user-a", created.ID))
	_, err = svc.Get("user-a", created.ID)
	require.ErrorIs(t, err, errConversationNotFound)
}

func TestConversationPinSortsFirst(t *testing.T) {
	svc := setupConversationService(t)

	older, err := svc.Create("user-a", &CreateConversationRequest{
		Title: "older",
		Messages: []model.ConversationMessage{
			{ID: "m1", Role: "user", Content: "older"},
		},
	})
	require.NoError(t, err)

	newer, err := svc.Create("user-a", &CreateConversationRequest{
		Title: "newer",
		Messages: []model.ConversationMessage{
			{ID: "m2", Role: "user", Content: "newer"},
		},
	})
	require.NoError(t, err)

	listed, err := svc.List("user-a")
	require.NoError(t, err)
	require.Len(t, listed, 2)
	assert.Equal(t, newer.ID, listed[0].ID)
	assert.False(t, listed[0].Pinned)

	pinned := true
	_, err = svc.Update("user-a", older.ID, &UpdateConversationRequest{Pinned: &pinned})
	require.NoError(t, err)

	listed, err = svc.List("user-a")
	require.NoError(t, err)
	require.Len(t, listed, 2)
	assert.Equal(t, older.ID, listed[0].ID)
	assert.True(t, listed[0].Pinned)
	assert.Equal(t, newer.ID, listed[1].ID)
	assert.False(t, listed[1].Pinned)

	unpinned := false
	_, err = svc.Update("user-a", older.ID, &UpdateConversationRequest{Pinned: &unpinned})
	require.NoError(t, err)

	listed, err = svc.List("user-a")
	require.NoError(t, err)
	require.Len(t, listed, 2)
	assert.False(t, listed[0].Pinned)
	assert.False(t, listed[1].Pinned)
}

func strPtr(s string) *string { return &s }
