package team

import (
	"fmt"
	"strings"
	"time"

	"cogniforge/internal/model"
)

// HandoffRecord 一次显式交棒（写入 RunState，可进 checkpoint）
type HandoffRecord struct {
	FromAgentID string    `json:"from_agent_id"`
	ToAgentID   string    `json:"to_agent_id"`
	Reason      string    `json:"reason,omitempty"`
	At          time.Time `json:"at"`
}

// DelegateRecord 主管委派工人的一次记录
type DelegateRecord struct {
	WorkerAgentID string    `json:"worker_agent_id"`
	Task          string    `json:"task,omitempty"`
	ResultSummary string    `json:"result_summary,omitempty"`
	At            time.Time `json:"at"`
}

// RunState 小队一次 run 的共享状态（对齐设计 §6.2；可塞进 checkpoint payload）
type RunState struct {
	TeamID          string           `json:"team_id"`
	ActiveAgentID   string           `json:"active_agent_id"`
	Goal            string           `json:"goal,omitempty"`
	Scratchpad      map[string]any   `json:"scratchpad,omitempty"`
	Artifacts       []any            `json:"artifacts,omitempty"`
	HandoffHistory  []HandoffRecord  `json:"handoff_history,omitempty"`
	DelegateHistory []DelegateRecord `json:"delegate_history,omitempty"`
	StepCount       int              `json:"step_count"`
	MaxDepth        int              `json:"max_depth"`
	MaxSteps        int              `json:"max_steps"`
}

// ResumeRunState 续跑：校验 active 属于小队后沿用历史；非法则回落 Init
func ResumeRunState(t *model.AgentTeam, prior *RunState, goal string) (*RunState, error) {
	if prior == nil || strings.TrimSpace(prior.ActiveAgentID) == "" {
		return InitRunState(t, goal)
	}
	if t == nil {
		return nil, fmt.Errorf("team is nil")
	}
	if prior.TeamID != "" && prior.TeamID != t.ID {
		return InitRunState(t, goal)
	}
	if !teamAllowsAgent(t, prior.ActiveAgentID) {
		return InitRunState(t, goal)
	}
	out := *prior
	out.TeamID = t.ID
	if out.Scratchpad == nil {
		out.Scratchpad = map[string]any{}
	}
	if out.MaxDepth <= 0 {
		out.MaxDepth = 8
	}
	if out.MaxSteps <= 0 {
		out.MaxSteps = 40
	}
	if goal != "" {
		out.Goal = goal
	}
	return &out, nil
}

// NewRunState 从小队配置构造初始状态
func NewRunState(teamID, activeAgentID, goal string, maxDepth, maxSteps int) *RunState {
	if maxDepth <= 0 {
		maxDepth = 8
	}
	if maxSteps <= 0 {
		maxSteps = 40
	}
	return &RunState{
		TeamID:          teamID,
		ActiveAgentID:   activeAgentID,
		Goal:            goal,
		Scratchpad:      map[string]any{},
		Artifacts:       []any{},
		HandoffHistory:  []HandoffRecord{},
		DelegateHistory: []DelegateRecord{},
		MaxDepth:        maxDepth,
		MaxSteps:        maxSteps,
	}
}

// CanStep 是否还能再走一步 LLM/委派
func (s *RunState) CanStep() error {
	if s == nil {
		return fmt.Errorf("run state is nil")
	}
	if s.StepCount >= s.MaxSteps {
		return fmt.Errorf("team run exceeded max_steps (%d)", s.MaxSteps)
	}
	return nil
}

// IncStep 成功走完一步后调用
func (s *RunState) IncStep() {
	if s != nil {
		s.StepCount++
	}
}

// Handoff 显式交棒；校验深度与连续来回
func (s *RunState) Handoff(toAgentID, reason string) error {
	if s == nil {
		return fmt.Errorf("run state is nil")
	}
	if toAgentID == "" {
		return fmt.Errorf("handoff: to_agent_id required")
	}
	if toAgentID == s.ActiveAgentID {
		return fmt.Errorf("handoff: already active agent %s", toAgentID)
	}
	if err := s.CanStep(); err != nil {
		return err
	}
	if len(s.HandoffHistory) >= s.MaxDepth {
		return fmt.Errorf("handoff: exceeded max_depth (%d)", s.MaxDepth)
	}
	// A→B→A→B 式来回：当前边若是对上一条的反转，且再前一条与当前同向，则打断
	if n := len(s.HandoffHistory); n >= 2 {
		last := s.HandoffHistory[n-1]
		prev := s.HandoffHistory[n-2]
		reversing := last.ToAgentID == s.ActiveAgentID && last.FromAgentID == toAgentID
		sameAsPrev := prev.FromAgentID == s.ActiveAgentID && prev.ToAgentID == toAgentID
		if reversing && sameAsPrev {
			return fmt.Errorf("handoff: ping-pong %s↔%s blocked", s.ActiveAgentID, toAgentID)
		}
	}
	from := s.ActiveAgentID
	s.HandoffHistory = append(s.HandoffHistory, HandoffRecord{
		FromAgentID: from,
		ToAgentID:   toAgentID,
		Reason:      reason,
		At:          time.Now().UTC(),
	})
	s.ActiveAgentID = toAgentID
	s.IncStep()
	return nil
}

// RecordDelegate 记录一次委派（深度计入 handoff+delegate 总量上限用 MaxDepth 近似）
func (s *RunState) RecordDelegate(workerID, task, resultSummary string) error {
	if s == nil {
		return fmt.Errorf("run state is nil")
	}
	if workerID == "" {
		return fmt.Errorf("delegate: worker_agent_id required")
	}
	if err := s.CanStep(); err != nil {
		return err
	}
	depth := len(s.HandoffHistory) + len(s.DelegateHistory)
	if depth >= s.MaxDepth {
		return fmt.Errorf("delegate: exceeded max_depth (%d)", s.MaxDepth)
	}
	s.DelegateHistory = append(s.DelegateHistory, DelegateRecord{
		WorkerAgentID: workerID,
		Task:          task,
		ResultSummary: resultSummary,
		At:            time.Now().UTC(),
	})
	s.IncStep()
	return nil
}

// SetScratch 写共享草稿
func (s *RunState) SetScratch(key string, value any) {
	if s == nil {
		return
	}
	if s.Scratchpad == nil {
		s.Scratchpad = map[string]any{}
	}
	s.Scratchpad[key] = value
}
