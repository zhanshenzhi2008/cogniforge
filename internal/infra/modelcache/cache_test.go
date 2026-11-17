package modelcache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocalPlusRemoteRevConsistency(t *testing.T) {
	remote := NewMemoryRemote()
	a := New(remote)
	b := New(remote)

	snap := &Snapshot{Rev: 1, ID: "deepseek", DefaultModel: "deepseek-chat", EncryptedKey: "enc"}
	require.NoError(t, remote.Save(nil, snap))

	got, ok := a.Get()
	require.True(t, ok)
	require.Equal(t, "deepseek-chat", got.DefaultModel)

	// 另一进程本地也应从 Redis 读到同一份
	gotB, ok := b.Get()
	require.True(t, ok)
	require.Equal(t, "deepseek", gotB.ID)

	a.PutHot(snap, "sk-plain", map[string]string{"X": "1"})
	_, plain, hdrs, ok := a.GetHot()
	require.True(t, ok)
	require.Equal(t, "sk-plain", plain)
	require.Equal(t, "1", hdrs["X"])

	// 配置变更：rev+1，删 snapshot → 两边本地都失效
	a.Invalidate()
	_, ok = a.Get()
	require.False(t, ok)
	_, ok = b.Get()
	require.False(t, ok)

	fresh := &Snapshot{Rev: remote.rev, ID: "openai", DefaultModel: "gpt-4o"}
	a.Put(fresh)
	got, ok = b.Get()
	require.True(t, ok)
	require.Equal(t, "gpt-4o", got.DefaultModel)
}

func TestNilCacheSafe(t *testing.T) {
	var c *Cache
	_, ok := c.Get()
	require.False(t, ok)
	c.Invalidate()
	c.Put(&Snapshot{ID: "x"})
}
