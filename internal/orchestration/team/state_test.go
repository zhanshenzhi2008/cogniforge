package team

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandoffAndDepth(t *testing.T) {
	s := NewRunState("team-1", "a", "goal", 2, 10)
	require.NoError(t, s.Handoff("b", "to worker"))
	require.Equal(t, "b", s.ActiveAgentID)
	require.Len(t, s.HandoffHistory, 1)

	require.NoError(t, s.Handoff("c", "next"))
	require.Error(t, s.Handoff("a", "over depth"))
}

func TestHandoffPingPongBlocked(t *testing.T) {
	s := NewRunState("team-1", "a", "goal", 8, 20)
	require.NoError(t, s.Handoff("b", "1"))
	require.NoError(t, s.Handoff("a", "2"))
	err := s.Handoff("b", "3 pingpong")
	require.Error(t, err)
	require.Contains(t, err.Error(), "ping-pong")
}

func TestMaxSteps(t *testing.T) {
	s := NewRunState("team-1", "a", "goal", 8, 2)
	require.NoError(t, s.RecordDelegate("w1", "task", "ok"))
	require.NoError(t, s.RecordDelegate("w2", "task", "ok"))
	require.Error(t, s.RecordDelegate("w3", "task", "ok"))
}

func TestSelfHandoffRejected(t *testing.T) {
	s := NewRunState("team-1", "a", "goal", 8, 10)
	require.Error(t, s.Handoff("a", "noop"))
}
