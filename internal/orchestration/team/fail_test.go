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
