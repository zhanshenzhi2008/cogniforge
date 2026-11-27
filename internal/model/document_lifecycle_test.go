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
