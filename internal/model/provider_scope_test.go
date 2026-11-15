package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProviderVisibleTo(t *testing.T) {
	pub := &AIProvider{Scope: ProviderScopePublic, OwnerUserID: "admin1"}
	priv := &AIProvider{Scope: ProviderScopePrivate, OwnerUserID: "u1"}

	assert.True(t, ProviderVisibleTo(pub, "anyone", false))
	assert.True(t, ProviderVisibleTo(priv, "u1", false))
	assert.False(t, ProviderVisibleTo(priv, "u2", false))
	assert.True(t, ProviderVisibleTo(priv, "u2", true))
}

func TestProviderWritableBy(t *testing.T) {
	pub := &AIProvider{Scope: ProviderScopePublic, OwnerUserID: "admin1"}
	priv := &AIProvider{Scope: ProviderScopePrivate, OwnerUserID: "u1"}

	assert.False(t, ProviderWritableBy(pub, "u1", false))
	assert.True(t, ProviderWritableBy(pub, "admin1", true))
	assert.True(t, ProviderWritableBy(priv, "u1", false))
	assert.False(t, ProviderWritableBy(priv, "u2", false))
	assert.True(t, ProviderWritableBy(priv, "u2", true))
}
