package debug_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"cogniforge/internal/engine/debug"
)

func TestNewDebugger(t *testing.T) {
	d := debug.NewDebugger("exec-1", "wf-1")
	assert.NotNil(t, d)
	assert.Equal(t, "exec-1", d.GetStatus().ExecutionID)
	assert.Equal(t, "wf-1", d.GetStatus().WorkflowID)
	assert.Equal(t, debug.DebugModeAuto, d.GetMode())
	assert.Equal(t, debug.DebugStateRunning, d.GetState())
}

func TestDebugger_SetMode(t *testing.T) {
	d := debug.NewDebugger("exec-1", "wf-1")

	d.SetMode(debug.DebugModeStep)
	assert.Equal(t, debug.DebugModeStep, d.GetMode())

	d.SetMode(debug.DebugModePause)
	assert.Equal(t, debug.DebugModePause, d.GetMode())
}

func TestDebugger_SetState(t *testing.T) {
	d := debug.NewDebugger("exec-1", "wf-1")

	d.SetState(debug.DebugStatePaused)
	assert.Equal(t, debug.DebugStatePaused, d.GetState())

	d.SetState(debug.DebugStateStopped)
	assert.Equal(t, debug.DebugStateStopped, d.GetState())

	d.SetState(debug.DebugStateDone)
	assert.Equal(t, debug.DebugStateDone, d.GetState())
}

func TestDebugger_CurrentNode(t *testing.T) {
	d := debug.NewDebugger("exec-1", "wf-1")

	d.SetCurrentNode("node-1")
	assert.Equal(t, "node-1", d.GetCurrentNode())

	d.SetCurrentNode("node-2")
	assert.Equal(t, "node-2", d.GetCurrentNode())
}

func TestDebugger_Breakpoints(t *testing.T) {
	d := debug.NewDebugger("exec-1", "wf-1")

	d.AddBreakpoint("node-1", "")
	assert.True(t, d.HasBreakpoint("node-1"))

	d.RemoveBreakpoint("node-1")
	assert.False(t, d.HasBreakpoint("node-1"))

	d.AddBreakpoint("node-2", "count > 5")
	assert.True(t, d.HasBreakpoint("node-2"))

	d.DisableBreakpoint("node-2")
	assert.False(t, d.HasBreakpoint("node-2"))

	d.EnableBreakpoint("node-2")
	assert.True(t, d.HasBreakpoint("node-2"))
}

func TestDebugger_GetBreakpoints(t *testing.T) {
	d := debug.NewDebugger("exec-1", "wf-1")

	d.AddBreakpoint("node-1", "")
	d.AddBreakpoint("node-2", "count > 5")

	breakpoints := d.GetBreakpoints()
	assert.Len(t, breakpoints, 2)
	assert.NotNil(t, breakpoints["node-1"])
	assert.NotNil(t, breakpoints["node-2"])
}

func TestDebugger_WatchVars(t *testing.T) {
	d := debug.NewDebugger("exec-1", "wf-1")

	d.AddWatchVar("count")
	d.AddWatchVar("name")
	d.AddWatchVar("status")

	watchVars := d.GetWatchVars()
	assert.Len(t, watchVars, 3)
	assert.Contains(t, watchVars, "count")
	assert.Contains(t, watchVars, "name")
	assert.Contains(t, watchVars, "status")

	d.RemoveWatchVar("name")
	watchVars = d.GetWatchVars()
	assert.Len(t, watchVars, 2)
	assert.NotContains(t, watchVars, "name")
}

func TestDebugger_StepCount(t *testing.T) {
	d := debug.NewDebugger("exec-1", "wf-1")

	assert.Equal(t, 0, d.GetStepCount())

	d.IncrementStep()
	d.IncrementStep()
	d.IncrementStep()

	assert.Equal(t, 3, d.GetStepCount())
}

func TestDebugger_ShouldPause(t *testing.T) {
	d := debug.NewDebugger("exec-1", "wf-1")

	assert.False(t, d.ShouldPause("node-1"))

	d.AddBreakpoint("node-1", "")
	assert.True(t, d.ShouldPause("node-1"))

	d.RemoveBreakpoint("node-1")
	d.SetMode(debug.DebugModeStep)
	assert.True(t, d.ShouldPause("node-2"))

	d.SetMode(debug.DebugModeResume)
	assert.False(t, d.ShouldPause("node-3"))

	d.SetState(debug.DebugStatePaused)
	assert.True(t, d.ShouldPause("node-4"))
}

func TestDebugger_ProcessCommand(t *testing.T) {
	d := debug.NewDebugger("exec-1", "wf-1")

	tests := []struct {
		name        string
		cmd         debug.DebugCommand
		expectError bool
		checkFunc   func(*debug.Debugger)
	}{
		{
			name: "step command",
			cmd:  debug.DebugCommand{Type: "step"},
			checkFunc: func(d *debug.Debugger) {
				assert.Equal(t, debug.DebugModeStep, d.GetMode())
				assert.Equal(t, 1, d.GetStepCount())
			},
		},
		{
			name: "resume command",
			cmd:  debug.DebugCommand{Type: "resume"},
			checkFunc: func(d *debug.Debugger) {
				assert.Equal(t, debug.DebugModeResume, d.GetMode())
				assert.Equal(t, debug.DebugStateRunning, d.GetState())
			},
		},
		{
			name: "pause command",
			cmd:  debug.DebugCommand{Type: "pause"},
			checkFunc: func(d *debug.Debugger) {
				assert.Equal(t, debug.DebugStatePaused, d.GetState())
			},
		},
		{
			name: "stop command",
			cmd:  debug.DebugCommand{Type: "stop"},
			checkFunc: func(d *debug.Debugger) {
				assert.Equal(t, debug.DebugStateStopped, d.GetState())
			},
		},
		{
			name: "add breakpoint",
			cmd:  debug.DebugCommand{Type: "add_breakpoint", NodeID: "node-1"},
			checkFunc: func(d *debug.Debugger) {
				assert.True(t, d.HasBreakpoint("node-1"))
			},
		},
		{
			name: "remove breakpoint",
			cmd:  debug.DebugCommand{Type: "remove_breakpoint", NodeID: "node-1"},
			checkFunc: func(d *debug.Debugger) {
				assert.False(t, d.HasBreakpoint("node-1"))
			},
		},
		{
			name:        "unknown command",
			cmd:         debug.DebugCommand{Type: "unknown"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := d.ProcessCommand(tt.cmd)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(d)
				}
			}
		})
	}
}

func TestDebugger_GetStatus(t *testing.T) {
	d := debug.NewDebugger("exec-1", "wf-1")

	d.SetMode(debug.DebugModeStep)
	d.SetState(debug.DebugStatePaused)
	d.SetCurrentNode("node-3")
	d.AddBreakpoint("node-1", "")
	d.AddWatchVar("count")
	d.IncrementStep()
	d.IncrementStep()

	status := d.GetStatus()

	assert.Equal(t, "exec-1", status.ExecutionID)
	assert.Equal(t, "wf-1", status.WorkflowID)
	assert.Equal(t, debug.DebugModeStep, status.Mode)
	assert.Equal(t, debug.DebugStatePaused, status.State)
	assert.Equal(t, "node-3", status.CurrentNode)
	assert.Equal(t, 2, status.StepCount)
	assert.Len(t, status.Breakpoints, 1)
	assert.Contains(t, status.WatchVars, "count")
}

func TestBreakpoint_Struct(t *testing.T) {
	bp := debug.Breakpoint{
		NodeID:    "node-1",
		Condition: "count > 5",
		Enabled:   true,
	}

	assert.Equal(t, "node-1", bp.NodeID)
	assert.Equal(t, "count > 5", bp.Condition)
	assert.True(t, bp.Enabled)
}

func TestDebugMode_Constants(t *testing.T) {
	assert.Equal(t, debug.DebugMode("step"), debug.DebugModeStep)
	assert.Equal(t, debug.DebugMode("resume"), debug.DebugModeResume)
	assert.Equal(t, debug.DebugMode("pause"), debug.DebugModePause)
	assert.Equal(t, debug.DebugMode("stop"), debug.DebugModeStop)
	assert.Equal(t, debug.DebugMode("auto"), debug.DebugModeAuto)
}

func TestDebugState_Constants(t *testing.T) {
	assert.Equal(t, debug.DebugState("running"), debug.DebugStateRunning)
	assert.Equal(t, debug.DebugState("paused"), debug.DebugStatePaused)
	assert.Equal(t, debug.DebugState("stopped"), debug.DebugStateStopped)
	assert.Equal(t, debug.DebugState("done"), debug.DebugStateDone)
}
