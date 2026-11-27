package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDocument_IsSearchableNow(t *testing.T) {
	now := time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	assert.True(t, (&Document{LifecycleStatus: DocLifecyclePublished}).IsSearchableNow(now))
	assert.True(t, (&Document{}).IsSearchableNow(now), "empty lifecycle treats as published")
	assert.False(t, (&Document{LifecycleStatus: DocLifecycleDraft}).IsSearchableNow(now))
	assert.False(t, (&Document{LifecycleStatus: DocLifecycleReview}).IsSearchableNow(now))
	assert.False(t, (&Document{LifecycleStatus: DocLifecycleAbolished}).IsSearchableNow(now))

	assert.False(t, (&Document{
		LifecycleStatus: DocLifecyclePublished,
		EffectiveAt:     &future,
	}).IsSearchableNow(now))
	assert.True(t, (&Document{
		LifecycleStatus: DocLifecyclePublished,
		EffectiveAt:     &past,
	}).IsSearchableNow(now))
	assert.False(t, (&Document{
		LifecycleStatus: DocLifecyclePublished,
		ExpireAt:        &past,
	}).IsSearchableNow(now))
	assert.True(t, (&Document{
		LifecycleStatus: DocLifecyclePublished,
		ExpireAt:        &future,
	}).IsSearchableNow(now))
}

func TestSecurityLevelRankAndClearance(t *testing.T) {
	assert.Equal(t, 1, SecurityLevelRank(DocSecurityPublic))
	assert.Equal(t, 2, SecurityLevelRank(DocSecurityInternal))
	assert.Equal(t, 3, SecurityLevelRank(DocSecurityConfidential))
	assert.Equal(t, 4, SecurityLevelRank(DocSecuritySecret))
	assert.Equal(t, 2, SecurityLevelRank("unknown"))

	doc := &Document{SecurityLevel: DocSecurityConfidential}
	assert.False(t, doc.AccessibleByClearance(DocSecurityInternal))
	assert.True(t, doc.AccessibleByClearance(DocSecurityConfidential))
	assert.True(t, doc.AccessibleByClearance(DocSecuritySecret))
	assert.True(t, (&Document{SecurityLevel: DocSecurityPublic}).AccessibleByClearance(DocSecurityInternal))
}
