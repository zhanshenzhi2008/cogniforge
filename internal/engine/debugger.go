package engine

import (
	"fmt"
	"sync"
)

type DebugMode string

const (
	DebugModeStep   DebugMode = "step"   // 单步执行
	DebugModeResume DebugMode = "resume" // 继续执行
	DebugModePause  DebugMode = "pause"  // 暂停
	DebugModeStop   DebugMode = "stop"   // 停止
	DebugModeAuto   DebugMode = "auto"   // 自动执行（无断点）
)

type Breakpoint struct {
	NodeID    string `json:"node_id"`
	Condition string `json:"condition,omitempty"` // 条件断点
	Enabled   bool   `json:"enabled"`
}

type DebugState string

const (
	DebugStateRunning DebugState = "running"
	DebugStatePaused  DebugState = "paused"
	DebugStateStopped DebugState = "stopped"
	DebugStateDone    DebugState = "done"
)

type Debugger struct {
	executionID string
	workflowID  string
	mode        DebugMode
	state       DebugState
	breakpoints map[string]*Breakpoint
	currentNode string
	stepCount   int
	watchVars   []string
	mu          sync.RWMutex
	pauseChan   chan struct{}
	resumeChan  chan struct{}
	stopChan    chan struct{}
}

func NewDebugger(executionID, workflowID string) *Debugger {
	return &Debugger{
		executionID: executionID,
		workflowID:  workflowID,
		mode:        DebugModeAuto,
		state:       DebugStateRunning,
		breakpoints: make(map[string]*Breakpoint),
		watchVars:   make([]string, 0),
		pauseChan:   make(chan struct{}),
		resumeChan:  make(chan struct{}),
		stopChan:    make(chan struct{}),
	}
}

func (d *Debugger) SetMode(mode DebugMode) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.mode = mode
}

func (d *Debugger) GetMode() DebugMode {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.mode
}

func (d *Debugger) GetState() DebugState {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.state
}

func (d *Debugger) SetState(state DebugState) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.state = state
}

func (d *Debugger) SetCurrentNode(nodeID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.currentNode = nodeID
}

func (d *Debugger) GetCurrentNode() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.currentNode
}

func (d *Debugger) AddBreakpoint(nodeID string, condition string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.breakpoints[nodeID] = &Breakpoint{
		NodeID:    nodeID,
		Condition: condition,
		Enabled:   true,
	}
}

func (d *Debugger) RemoveBreakpoint(nodeID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.breakpoints, nodeID)
}

func (d *Debugger) EnableBreakpoint(nodeID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if bp, ok := d.breakpoints[nodeID]; ok {
		bp.Enabled = true
	}
}

func (d *Debugger) DisableBreakpoint(nodeID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if bp, ok := d.breakpoints[nodeID]; ok {
		bp.Enabled = false
	}
}

func (d *Debugger) GetBreakpoints() map[string]*Breakpoint {
	d.mu.RLock()
	defer d.mu.RUnlock()
	result := make(map[string]*Breakpoint)
	for k, v := range d.breakpoints {
		result[k] = v
	}
	return result
}

func (d *Debugger) HasBreakpoint(nodeID string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if bp, ok := d.breakpoints[nodeID]; ok {
		return bp.Enabled
	}
	return false
}

func (d *Debugger) AddWatchVar(variableName string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.watchVars = append(d.watchVars, variableName)
}

func (d *Debugger) RemoveWatchVar(variableName string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i, v := range d.watchVars {
		if v == variableName {
			d.watchVars = append(d.watchVars[:i], d.watchVars[i+1:]...)
			break
		}
	}
}

func (d *Debugger) GetWatchVars() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	result := make([]string, len(d.watchVars))
	copy(result, d.watchVars)
	return result
}

func (d *Debugger) IncrementStep() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.stepCount++
}

func (d *Debugger) GetStepCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.stepCount
}

func (d *Debugger) Pause() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.state = DebugStatePaused
	close(d.pauseChan)
	d.pauseChan = make(chan struct{})
}

func (d *Debugger) Resume() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.state = DebugStateRunning
	d.resumeChan = make(chan struct{})
}

func (d *Debugger) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.state = DebugStateStopped
	close(d.stopChan)
}

func (d *Debugger) ShouldPause(nodeID string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Check if paused
	if d.state == DebugStatePaused {
		return true
	}

	// Check for breakpoint at this node
	if bp, ok := d.breakpoints[nodeID]; ok && bp.Enabled {
		return true
	}

	// Check if in step mode
	if d.mode == DebugModeStep {
		return true
	}

	return false
}

func (d *Debugger) WaitForResume() {
	<-d.resumeChan
}

func (d *Debugger) WaitForStop() bool {
	select {
	case <-d.stopChan:
		return true
	default:
		return false
	}
}

func (d *Debugger) GetStatus() DebugStatus {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return DebugStatus{
		ExecutionID: d.executionID,
		WorkflowID:  d.workflowID,
		Mode:        d.mode,
		State:       d.state,
		CurrentNode: d.currentNode,
		StepCount:   d.stepCount,
		Breakpoints: d.GetBreakpoints(),
		WatchVars:   d.watchVars,
	}
}

type DebugStatus struct {
	ExecutionID string                 `json:"execution_id"`
	WorkflowID  string                 `json:"workflow_id"`
	Mode        DebugMode              `json:"mode"`
	State       DebugState             `json:"state"`
	CurrentNode string                 `json:"current_node"`
	StepCount   int                    `json:"step_count"`
	Breakpoints map[string]*Breakpoint `json:"breakpoints"`
	WatchVars   []string               `json:"watch_vars"`
}

type DebugCommand struct {
	Type      string `json:"type"` // step, resume, pause, stop, add_breakpoint, remove_breakpoint
	NodeID    string `json:"node_id,omitempty"`
	Condition string `json:"condition,omitempty"`
}

func (d *Debugger) ProcessCommand(cmd DebugCommand) error {
	switch cmd.Type {
	case "step":
		d.SetMode(DebugModeStep)
		d.IncrementStep()
		d.Resume()
	case "resume":
		d.SetMode(DebugModeResume)
		d.Resume()
	case "pause":
		d.Pause()
	case "stop":
		d.Stop()
	case "add_breakpoint":
		d.AddBreakpoint(cmd.NodeID, cmd.Condition)
	case "remove_breakpoint":
		d.RemoveBreakpoint(cmd.NodeID)
	case "enable_breakpoint":
		d.EnableBreakpoint(cmd.NodeID)
	case "disable_breakpoint":
		d.DisableBreakpoint(cmd.NodeID)
	default:
		return fmt.Errorf("unknown debug command: %s", cmd.Type)
	}
	return nil
}
