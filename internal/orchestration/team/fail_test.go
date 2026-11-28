package team

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cogniforge/internal/model"
)

func TestWorkerRetryOnceConfig(t *testing.T) {
	require.False(t, workerRetryOnce(nil))
	require.False(t, workerRetryOnce(&model.AgentTeam{}))
	require.False(t, workerRetryOnce(&model.AgentTeam{Config: model.JSONBMap{"worker_retry_once": false}}))
	require.True(t, workerRetryOnce(&model.AgentTeam{Config: model.JSONBMap{"worker_retry_once": true}}))
	require.True(t, workerRetryOnce(&model.AgentTeam{Config: model.JSONBMap{"worker_retry_once": "true"}}))
}

func TestWorkerAutoHandoffConfig(t *testing.T) {
	require.False(t, workerAutoHandoff(nil))
	require.False(t, workerAutoHandoff(&model.AgentTeam{}))
	require.True(t, workerAutoHandoff(&model.AgentTeam{Config: model.JSONBMap{"worker_auto_handoff": true}}))
	require.True(t, workerAutoHandoff(&model.AgentTeam{Config: model.JSONBMap{"worker_auto_handoff": "1"}}))
}

func TestWorkerFallbackIDs(t *testing.T) {
	require.Nil(t, workerFallbackIDs(nil))
	require.Nil(t, workerFallbackIDs(&model.AgentTeam{}))
	require.Equal(t, []string{"a", "b"}, workerFallbackIDs(&model.AgentTeam{
		Config: model.JSONBMap{"worker_fallback_agent_ids": []string{"a", "b"}},
	}))
	require.Equal(t, []string{"x", "y"}, workerFallbackIDs(&model.AgentTeam{
		Config: model.JSONBMap{"worker_fallback_agent_ids": []any{"x", " y "}},
	}))
	require.Equal(t, []string{"p", "q"}, workerFallbackIDs(&model.AgentTeam{
		Config: model.JSONBMap{"worker_fallback_agent_ids": "p, q"},
	}))
}
