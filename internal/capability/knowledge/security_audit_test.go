package knowledge

import (
	"testing"

	"cogniforge/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestWriteAndListSecurityAudits(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:secaudit?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.SecurityAuditLog{}))

	s := &KnowledgeService{db: db}
	s.writeSecurityAudit("u1", "document_security_level_change", "document", "d1", "kb1",
		model.DocSecurityInternal, model.DocSecurityConfidential, "success", "")
	s.writeSecurityAudit("u2", "document_security_level_change", "document", "d2", "kb1",
		model.DocSecurityInternal, model.DocSecuritySecret, "denied", "无权")

	rows, err := s.ListSecurityAudits(10)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, "denied", rows[0].Result)
	assert.Equal(t, "success", rows[1].Result)
	assert.Equal(t, "d2", rows[0].ResourceID)
}
