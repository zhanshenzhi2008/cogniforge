package team

import (
	"fmt"
	"strings"

	"cogniforge/internal/model"
)

// BuildLinearPipelineDefinition 从小队成员生成 start→agent…→end 线性图（Vue Flow 形态）。
// agentNames：id → 展示名；缺省用 id。
func BuildLinearPipelineDefinition(t *model.AgentTeam, agentNames map[string]string) (map[string]any, error) {
	if t == nil {
		return nil, fmt.Errorf("team is nil")
	}
	ids := pipelineAgentOrder(t)
	if len(ids) == 0 {
		return nil, fmt.Errorf("小队没有可串联的成员 Agent")
	}

	const x0, y0, dx = 80.0, 160.0, 220.0
	nodes := make([]map[string]any, 0, len(ids)+2)
	edges := make([]map[string]any, 0, len(ids)+1)

	startID := "start"
	nodes = append(nodes, map[string]any{
		"id":       startID,
		"type":     "start",
		"name":     "开始",
		"position": map[string]any{"x": x0, "y": y0},
		"data":     map[string]any{"label": "开始"},
	})

	prev := startID
	for i, aid := range ids {
		nid := fmt.Sprintf("agent-%d", i+1)
		label := aid
		if agentNames != nil {
			if n := strings.TrimSpace(agentNames[aid]); n != "" {
				label = n
			}
		}
		nodes = append(nodes, map[string]any{
			"id":       nid,
			"type":     "agent",
			"name":     label,
			"position": map[string]any{"x": x0 + float64(i+1)*dx, "y": y0},
			"data": map[string]any{
				"label":   label,
				"agentId": aid,
				"prompt":  "", // 空：首节点吃 input.query，后续吃上一步 content（引擎变量）
			},
			"config": map[string]any{
				"agent_id": aid,
			},
		})
		edges = append(edges, map[string]any{
			"id":     fmt.Sprintf("e-%s-%s", prev, nid),
			"source": prev,
			"target": nid,
		})
		prev = nid
	}

	endID := "end"
	nodes = append(nodes, map[string]any{
		"id":       endID,
		"type":     "end",
		"name":     "结束",
		"position": map[string]any{"x": x0 + float64(len(ids)+1)*dx, "y": y0},
		"data":     map[string]any{"label": "结束"},
	})
	edges = append(edges, map[string]any{
		"id":     fmt.Sprintf("e-%s-%s", prev, endID),
		"source": prev,
		"target": endID,
	})

	return map[string]any{
		"nodes": nodes,
		"edges": edges,
	}, nil
}

// BuildTeamNodeDefinition supervisor/handoff：start→team→end
func BuildTeamNodeDefinition(t *model.AgentTeam) (map[string]any, error) {
	if t == nil || strings.TrimSpace(t.ID) == "" {
		return nil, fmt.Errorf("team is nil")
	}
	label := strings.TrimSpace(t.Name)
	if label == "" {
		label = "小队"
	}
	return map[string]any{
		"nodes": []map[string]any{
			{
				"id": "start", "type": "start", "name": "开始",
				"position": map[string]any{"x": 80, "y": 160},
				"data":     map[string]any{"label": "开始"},
			},
			{
				"id": "team-1", "type": "team", "name": label,
				"position": map[string]any{"x": 300, "y": 160},
				"data": map[string]any{
					"label":  label,
					"teamId": t.ID,
					"prompt": "",
				},
				"config": map[string]any{
					"team_id": t.ID,
					"task":    "",
				},
			},
			{
				"id": "end", "type": "end", "name": "结束",
				"position": map[string]any{"x": 520, "y": 160},
				"data":     map[string]any{"label": "结束"},
			},
		},
		"edges": []map[string]any{
			{"id": "e-start-team", "source": "start", "target": "team-1"},
			{"id": "e-team-end", "source": "team-1", "target": "end"},
		},
	}, nil
}

func pipelineAgentOrder(t *model.AgentTeam) []string {
	seen := map[string]struct{}{}
	var out []string
	add := func(id string) {
		id = strings.TrimSpace(id)
		if id == "" {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	// pipeline：成员顺序即流水线；若设了主管且不在列表里则放最前
	if t.Mode == model.TeamModePipeline && t.SupervisorAgentID != "" {
		add(t.SupervisorAgentID)
	}
	for _, id := range t.MemberAgentIDs {
		add(id)
	}
	if t.Mode != model.TeamModePipeline && len(out) == 0 && t.SupervisorAgentID != "" {
		add(t.SupervisorAgentID)
	}
	return out
}
