package engine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"cogniforge/internal/engine"
)

func TestNewDebugger(t *testing.T) {
	d := engine.NewDebugger("exec-1", "wf-1")
	assert.NotNil(t, d)
	assert.Equal(t, "exec-1", d.GetStatus().ExecutionID)
	assert.Equal(t, "wf-1", d.GetStatus().WorkflowID)
	assert.Equal(t, engine.DebugModeAuto, d.GetMode())
	assert.Equal(t, engine.DebugStateRunning, d.GetState())
}

func TestDebugger_SetMode(t *testing.T) {
	d := engine.NewDebugger("exec-1", "wf-1")

	d.SetMode(engine.DebugModeStep)
	assert.Equal(t, engine.DebugModeStep, d.GetMode())

	d.SetMode(engine.DebugModePause)
	assert.Equal(t, engine.DebugModePause, d.GetMode())
}

func TestDebugger_SetState(t *testing.T) {
	d := engine.NewDebugger("exec-1", "wf-1")

	d.SetState(engine.DebugStatePaused)
	assert.Equal(t, engine.DebugStatePaused, d.GetState())

	d.SetState(engine.DebugStateStopped)
	assert.Equal(t, engine.DebugStateStopped, d.GetState())

	d.SetState(engine.DebugStateDone)
	assert.Equal(t, engine.DebugStateDone, d.GetState())
}

func TestDebugger_CurrentNode(t *testing.T) {
	d := engine.NewDebugger("exec-1", "wf-1")

	d.SetCurrentNode("node-1")
	assert.Equal(t, "node-1", d.GetCurrentNode())

	d.SetCurrentNode("node-2")
	assert.Equal(t, "node-2", d.GetCurrentNode())
}

func TestDebugger_Breakpoints(t *testing.T) {
	d := engine.NewDebugger("exec-1", "wf-1")

	// Add breakpoint
	d.AddBreakpoint("node-1", "")
	assert.True(t, d.HasBreakpoint("node-1"))

	// Remove breakpoint
	d.RemoveBreakpoint("node-1")
	assert.False(t, d.HasBreakpoint("node-1"))

	// Add conditional breakpoint
	d.AddBreakpoint("node-2", "count > 5")
	assert.True(t, d.HasBreakpoint("node-2"))

	// Disable breakpoint
	d.DisableBreakpoint("node-2")
	assert.False(t, d.HasBreakpoint("node-2"))

	// Enable breakpoint
	d.EnableBreakpoint("node-2")
	assert.True(t, d.HasBreakpoint("node-2"))
}

func TestDebugger_GetBreakpoints(t *testing.T) {
	d := engine.NewDebugger("exec-1", "wf-1")

	d.AddBreakpoint("node-1", "")
	d.AddBreakpoint("node-2", "count > 5")

	breakpoints := d.GetBreakpoints()
	assert.Len(t, breakpoints, 2)
	assert.NotNil(t, breakpoints["node-1"])
	assert.NotNil(t, breakpoints["node-2"])
}

func TestDebugger_WatchVars(t *testing.T) {
	d := engine.NewDebugger("exec-1", "wf-1")

	// Add watch vars
	d.AddWatchVar("count")
	d.AddWatchVar("name")
	d.AddWatchVar("status")

	watchVars := d.GetWatchVars()
	assert.Len(t, watchVars, 3)
	assert.Contains(t, watchVars, "count")
	assert.Contains(t, watchVars, "name")
	assert.Contains(t, watchVars, "status")

	// Remove watch var
	d.RemoveWatchVar("name")
	watchVars = d.GetWatchVars()
	assert.Len(t, watchVars, 2)
	assert.NotContains(t, watchVars, "name")
}

func TestDebugger_StepCount(t *testing.T) {
	d := engine.NewDebugger("exec-1", "wf-1")

	assert.Equal(t, 0, d.GetStepCount())

	d.IncrementStep()
	d.IncrementStep()
	d.IncrementStep()

	assert.Equal(t, 3, d.GetStepCount())
}

func TestDebugger_ShouldPause(t *testing.T) {
	d := engine.NewDebugger("exec-1", "wf-1")

	// Not paused, no breakpoints - should not pause
	assert.False(t, d.ShouldPause("node-1"))

	// Add breakpoint
	d.AddBreakpoint("node-1", "")
	assert.True(t, d.ShouldPause("node-1"))

	// Remove breakpoint, set to step mode
	d.RemoveBreakpoint("node-1")
	d.SetMode(engine.DebugModeStep)
	assert.True(t, d.ShouldPause("node-2"))

	// Resume mode, no breakpoints
	d.SetMode(engine.DebugModeResume)
	assert.False(t, d.ShouldPause("node-3"))

	// Set to paused state
	d.SetState(engine.DebugStatePaused)
	assert.True(t, d.ShouldPause("node-4"))
}

func TestDebugger_ProcessCommand(t *testing.T) {
	d := engine.NewDebugger("exec-1", "wf-1")

	tests := []struct {
		name        string
		cmd         engine.DebugCommand
		expectError bool
		checkFunc   func(*engine.Debugger)
	}{
		{
			name: "step command",
			cmd:  engine.DebugCommand{Type: "step"},
			checkFunc: func(d *engine.Debugger) {
				assert.Equal(t, engine.DebugModeStep, d.GetMode())
				assert.Equal(t, 1, d.GetStepCount())
			},
		},
		{
			name: "resume command",
			cmd:  engine.DebugCommand{Type: "resume"},
			checkFunc: func(d *engine.Debugger) {
				assert.Equal(t, engine.DebugModeResume, d.GetMode())
				assert.Equal(t, engine.DebugStateRunning, d.GetState())
			},
		},
		{
			name: "pause command",
			cmd:  engine.DebugCommand{Type: "pause"},
			checkFunc: func(d *engine.Debugger) {
				assert.Equal(t, engine.DebugStatePaused, d.GetState())
			},
		},
		{
			name: "stop command",
			cmd:  engine.DebugCommand{Type: "stop"},
			checkFunc: func(d *engine.Debugger) {
				assert.Equal(t, engine.DebugStateStopped, d.GetState())
			},
		},
		{
			name: "add breakpoint",
			cmd:  engine.DebugCommand{Type: "add_breakpoint", NodeID: "node-1"},
			checkFunc: func(d *engine.Debugger) {
				assert.True(t, d.HasBreakpoint("node-1"))
			},
		},
		{
			name: "remove breakpoint",
			cmd:  engine.DebugCommand{Type: "remove_breakpoint", NodeID: "node-1"},
			checkFunc: func(d *engine.Debugger) {
				assert.False(t, d.HasBreakpoint("node-1"))
			},
		},
		{
			name:        "unknown command",
			cmd:         engine.DebugCommand{Type: "unknown"},
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
	d := engine.NewDebugger("exec-1", "wf-1")

	// Set some state
	d.SetMode(engine.DebugModeStep)
	d.SetState(engine.DebugStatePaused)
	d.SetCurrentNode("node-3")
	d.AddBreakpoint("node-1", "")
	d.AddWatchVar("count")
	d.IncrementStep()
	d.IncrementStep()

	status := d.GetStatus()

	assert.Equal(t, "exec-1", status.ExecutionID)
	assert.Equal(t, "wf-1", status.WorkflowID)
	assert.Equal(t, engine.DebugModeStep, status.Mode)
	assert.Equal(t, engine.DebugStatePaused, status.State)
	assert.Equal(t, "node-3", status.CurrentNode)
	assert.Equal(t, 2, status.StepCount)
	assert.Len(t, status.Breakpoints, 1)
	assert.Contains(t, status.WatchVars, "count")
}

func TestBreakpoint_Struct(t *testing.T) {
	bp := engine.Breakpoint{
		NodeID:    "node-1",
		Condition: "count > 5",
		Enabled:   true,
	}

	assert.Equal(t, "node-1", bp.NodeID)
	assert.Equal(t, "count > 5", bp.Condition)
	assert.True(t, bp.Enabled)
}

func TestDebugMode_Constants(t *testing.T) {
	assert.Equal(t, engine.DebugMode("step"), engine.DebugModeStep)
	assert.Equal(t, engine.DebugMode("resume"), engine.DebugModeResume)
	assert.Equal(t, engine.DebugMode("pause"), engine.DebugModePause)
	assert.Equal(t, engine.DebugMode("stop"), engine.DebugModeStop)
	assert.Equal(t, engine.DebugMode("auto"), engine.DebugModeAuto)
}

func TestDebugState_Constants(t *testing.T) {
	assert.Equal(t, engine.DebugState("running"), engine.DebugStateRunning)
	assert.Equal(t, engine.DebugState("paused"), engine.DebugStatePaused)
	assert.Equal(t, engine.DebugState("stopped"), engine.DebugStateStopped)
	assert.Equal(t, engine.DebugState("done"), engine.DebugStateDone)
}
