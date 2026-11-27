package team

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cogniforge/internal/model"
)

func TestBuildLinearPipelineDefinition(t *testing.T) {
	def, err := BuildLinearPipelineDefinition(&model.AgentTeam{
		Mode:           model.TeamModePipeline,
		MemberAgentIDs: model.JSONBArray{"a1", "a2"},
	}, map[string]string{"a1": "调研", "a2": "写作"})
	require.NoError(t, err)
	nodes := def["nodes"].([]map[string]any)
	edges := def["edges"].([]map[string]any)
	require.Len(t, nodes, 4) // start + 2 agents + end
	require.Len(t, edges, 3)
	require.Equal(t, "agent", nodes[1]["type"])
	require.Equal(t, "a1", nodes[1]["data"].(map[string]any)["agentId"])
}

func TestBuildLinearPipelineEmpty(t *testing.T) {
	_, err := BuildLinearPipelineDefinition(&model.AgentTeam{Mode: model.TeamModePipeline}, nil)
	require.Error(t, err)
}

func TestBuildTeamNodeDefinition(t *testing.T) {
	def, err := BuildTeamNodeDefinition(&model.AgentTeam{ID: "t1", Name: "调研队", Mode: model.TeamModeSupervisor})
	require.NoError(t, err)
	nodes := def["nodes"].([]map[string]any)
	require.Equal(t, "team", nodes[1]["type"])
	require.Equal(t, "t1", nodes[1]["data"].(map[string]any)["teamId"])
}

func TestBuildParallelFanoutDefinition(t *testing.T) {
	def, err := BuildParallelFanoutDefinition(&model.AgentTeam{
		SupervisorAgentID: "sup",
		MemberAgentIDs:    model.JSONBArray{"sup", "w1", "w2"},
	}, map[string]string{"w1": "工人甲", "w2": "工人乙"})
	require.NoError(t, err)
	nodes := def["nodes"].([]map[string]any)
	edges := def["edges"].([]map[string]any)
	require.Len(t, nodes, 4) // start + 2 workers + end（主管已剔除）
	require.Len(t, edges, 4) // start→each, each→end
	require.Equal(t, "agent", nodes[1]["type"])
	require.Equal(t, "w1", nodes[1]["data"].(map[string]any)["agentId"])
}

func TestBuildParallelFanoutEmpty(t *testing.T) {
	_, err := BuildParallelFanoutDefinition(&model.AgentTeam{
		SupervisorAgentID: "sup",
		MemberAgentIDs:    model.JSONBArray{"sup"},
	}, nil)
	require.Error(t, err)
}
