package knowledge

import (
	"testing"

	"cogniforge/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestAuthorizeSecurityLevelChange_Ranks(t *testing.T) {
	// 无 DB 时 resolveUserClearance 会落到 secret；这里只测「公开/内部不拦」分支
	s := &KnowledgeService{}
	assert.NoError(t, s.authorizeSecurityLevelChange("u", model.DocSecurityInternal, model.DocSecurityPublic))
	assert.NoError(t, s.authorizeSecurityLevelChange("u", model.DocSecurityPublic, model.DocSecurityInternal))
}

func TestSecurityLevelRank_RaiseNeedsHigher(t *testing.T) {
	assert.True(t, model.SecurityLevelRank(model.DocSecurityConfidential) > model.SecurityLevelRank(model.DocSecurityInternal))
	assert.True(t, model.SecurityLevelRank(model.DocSecuritySecret) > model.SecurityLevelRank(model.DocSecurityConfidential))
}
