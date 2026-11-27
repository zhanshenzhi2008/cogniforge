package core

import "strings"

// outEdge 带源句柄的出边（条件节点真=s-top / 假=s-bottom）
type outEdge struct {
	Target string
	Handle string
}

// normalizeConditionHandle 统一画布句柄与条件输出 branch
func normalizeConditionHandle(h string) string {
	h = strings.TrimSpace(strings.ToLower(h))
	switch h {
	case "s-top", "true", "yes", "1":
		return "s-top"
	case "s-bottom", "false", "no", "0", "default":
		return "s-bottom"
	default:
		return h
	}
}

// chosenConditionHandle 条件节点输出 → 应走的源句柄
func chosenConditionHandle(output any) string {
	m, ok := output.(map[string]any)
	if !ok {
		return "s-top"
	}
	if r, ok := m["result"].(bool); ok {
		if r {
			return "s-top"
		}
		return "s-bottom"
	}
	branch, _ := m["branch"].(string)
	n := normalizeConditionHandle(branch)
	if n == "s-bottom" {
		return "s-bottom"
	}
	return "s-top"
}

func conditionHandleMatches(edgeHandle, chosen string) bool {
	if strings.TrimSpace(edgeHandle) == "" {
		return true // 旧图无句柄：兼容双开
	}
	return normalizeConditionHandle(edgeHandle) == normalizeConditionHandle(chosen)
}

// releaseSuccessors 释放后继；条件节点按真/假句柄独占，未选边级联 skipped
func releaseSuccessors(
	nodeID, nodeType string,
	output any,
	succ map[string][]outEdge,
	inDegree map[string]int,
	completedSet map[string]bool,
	completed *[]string,
	ready *[]string,
	ctx *ExecutionContext,
) {
	outs := succ[nodeID]
	exclusive := nodeType == "condition"
	chosen := ""
	if exclusive {
		chosen = chosenConditionHandle(output)
	}
	for _, oe := range outs {
		if exclusive && !conditionHandleMatches(oe.Handle, chosen) {
			skipArrival(oe.Target, succ, inDegree, completedSet, completed, ctx)
			continue
		}
		inDegree[oe.Target]--
		if inDegree[oe.Target] == 0 && !completedSet[oe.Target] {
			*ready = append(*ready, oe.Target)
		}
	}
}

// skipArrival 未选分支到达：扣依赖；依赖清零则标 skipped 并继续向下
func skipArrival(
	id string,
	succ map[string][]outEdge,
	inDegree map[string]int,
	completedSet map[string]bool,
	completed *[]string,
	ctx *ExecutionContext,
) {
	inDegree[id]--
	if inDegree[id] > 0 {
		return
	}
	if completedSet[id] {
		return
	}
	completedSet[id] = true
	*completed = append(*completed, id)
	if ctx != nil {
		ctx.SetNodeState(id, "skipped")
		if ctx.OnNodeEnd != nil {
			ctx.OnNodeEnd(id, "skipped")
		}
		if ctx.Logger != nil {
			ctx.Logger.Logf(id, "info", "Skipped (untaken condition branch)")
		}
	}
	for _, oe := range succ[id] {
		skipArrival(oe.Target, succ, inDegree, completedSet, completed, ctx)
	}
}
