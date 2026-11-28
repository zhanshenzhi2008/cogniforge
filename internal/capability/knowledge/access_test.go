package knowledge

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"cogniforge/internal/model"
)

func TestNormalizeVisibility(t *testing.T) {
	assert.Equal(t, VisibilityPrivate, normalizeVisibility(""))
	assert.Equal(t, VisibilityPrivate, normalizeVisibility("private"))
	assert.Equal(t, VisibilityPrivate, normalizeVisibility("PUBLIC"))
	assert.Equal(t, VisibilityPrivate, normalizeVisibility("weird"))
	assert.Equal(t, VisibilityOrg, normalizeVisibility("org"))
	assert.Equal(t, VisibilityOrg, normalizeVisibility("ORG"))
	assert.Equal(t, VisibilityOrg, normalizeVisibility(" org "))
}

func TestKBVisibility_OrgCrossUser(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:kbvis?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.KnowledgeBase{}))

	svc := &KnowledgeService{db: db}
	ownerID := "owner1"
	otherID := "other1"

	orgKB, err := svc.CreateKnowledgeBase(ownerID, &CreateKBRequest{
		Name:       "Org Shared KB",
		Visibility: VisibilityOrg,
	})
	require.NoError(t, err)
	assert.Equal(t, VisibilityOrg, orgKB.Visibility)

	privKB, err := svc.CreateKnowledgeBase(ownerID, &CreateKBRequest{
		Name:       "Private KB",
		Visibility: VisibilityPrivate,
	})
	require.NoError(t, err)
	assert.Equal(t, VisibilityPrivate, privKB.Visibility)

	got, err := svc.GetKnowledgeBase(otherID, orgKB.ID)
	require.NoError(t, err)
	assert.Equal(t, orgKB.ID, got.ID)

	list, err := svc.ListKnowledgeBases(otherID)
	require.NoError(t, err)
	var sawOrg, sawPriv bool
	for _, kb := range list {
		if kb.ID == orgKB.ID {
			sawOrg = true
		}
		if kb.ID == privKB.ID {
			sawPriv = true
		}
	}
	assert.True(t, sawOrg, "org kb should appear in other user's list")
	assert.False(t, sawPriv, "private kb must not appear in other user's list")

	_, err = svc.GetKnowledgeBase(otherID, privKB.ID)
	assert.Error(t, err)

	_, err = svc.UpdateKnowledgeBase(otherID, orgKB.ID, &UpdateKBRequest{Name: "Hacked"})
	assert.Error(t, err)
	err = svc.DeleteKnowledgeBase(otherID, orgKB.ID)
	assert.Error(t, err)

	_, err = svc.loadKBForRead(otherID, orgKB.ID)
	require.NoError(t, err)
	_, err = svc.loadKBForWrite(otherID, orgKB.ID)
	assert.Error(t, err)

	updated, err := svc.UpdateKnowledgeBase(ownerID, orgKB.ID, &UpdateKBRequest{Visibility: VisibilityPrivate})
	require.NoError(t, err)
	assert.Equal(t, VisibilityPrivate, updated.Visibility)
	_, err = svc.GetKnowledgeBase(otherID, orgKB.ID)
	assert.Error(t, err)
}
