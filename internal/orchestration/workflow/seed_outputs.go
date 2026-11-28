package workflow

import (
	"cogniforge/internal/model"
	"cogniforge/internal/orchestration/engine"
)

// pickSeedOutputs 从种子执行的 node_runs 里，为指定节点挑最近一次有内容的 output
func pickSeedOutputs(runs []model.WorkflowNodeRun, nodeIDs []string) map[string]map[string]any {
	want := make(map[string]bool, len(nodeIDs))
	for _, id := range nodeIDs {
		if id != "" {
			want[id] = true
		}
	}
	out := make(map[string]map[string]any)
	for _, r := range runs {
		if !want[r.NodeID] {
			continue
		}
		if r.NodeType == "human_action" {
			continue
		}
		if r.Output == nil || len(r.Output) == 0 {
			continue
		}
		// 列表按时间升序时后写覆盖；无序也取有内容的即可
		cp := make(map[string]any, len(r.Output))
		for k, v := range r.Output {
			cp[k] = v
		}
		out[r.NodeID] = cp
	}
	return out
}

// applySeedOutputsToContext 把种子输出写入执行上下文
func applySeedOutputsToContext(ctx *engine.ExecutionContext, outputs map[string]map[string]any) int {
	if ctx == nil || len(outputs) == 0 {
		return 0
	}
	n := 0
	for id, o := range outputs {
		ctx.SetNodeOutput(id, o)
		n++
	}
	return n
}
