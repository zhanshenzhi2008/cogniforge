package team

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cogniforge/internal/model"
)

func TestResumeRunStateKeepsActive(t *testing.T) {
	tm := &model.AgentTeam{
		ID:                "t1",
		SupervisorAgentID: "sup",
		MemberAgentIDs:    model.JSONBArray{"w1", "w2"},
		Config:            model.TeamConfigDefaults(),
	}
	prior := NewRunState("t1", "sup", "g", 8, 40)
	require.NoError(t, prior.Handoff("w1", "expert"))
	got, err := ResumeRunState(tm, prior, "")
	require.NoError(t, err)
	require.Equal(t, "w1", got.ActiveAgentID)
	require.Len(t, got.HandoffHistory, 1)
}

func TestResumeRunStateRejectsForeignMember(t *testing.T) {
	tm := &model.AgentTeam{
		ID:                "t1",
		SupervisorAgentID: "sup",
		MemberAgentIDs:    model.JSONBArray{"w1"},
	}
	prior := &RunState{TeamID: "t1", ActiveAgentID: "outsider", MaxDepth: 8, MaxSteps: 40}
	got, err := ResumeRunState(tm, prior, "goal")
	require.NoError(t, err)
	require.Equal(t, "sup", got.ActiveAgentID)
}
