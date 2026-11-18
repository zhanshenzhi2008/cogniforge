package mcp

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodeExecute_PythonHello(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		if _, err2 := exec.LookPath("python"); err2 != nil {
			t.Skip("python not installed")
		}
	}
	e := NewBuiltInExecutor()
	res, err := e.Execute("code_execute_python", map[string]interface{}{
		"code":    "print('hello-cf')",
		"timeout": float64(10),
	})
	require.NoError(t, err)
	assert.Equal(t, "ok", res["status"])
	assert.Contains(t, res["output"], "hello-cf")
}

func TestCodeExecute_RejectEmpty(t *testing.T) {
	e := NewBuiltInExecutor()
	_, err := e.Execute("code_execute_python", map[string]interface{}{"code": "  "})
	require.Error(t, err)
}

func TestCodeExecute_Timeout(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		if _, err2 := exec.LookPath("python"); err2 != nil {
			t.Skip("python not installed")
		}
	}
	e := NewBuiltInExecutor()
	res, err := e.Execute("code_execute_python", map[string]interface{}{
		"code":    "import time\ntime.sleep(5)\nprint('late')",
		"timeout": float64(1),
	})
	require.NoError(t, err)
	assert.Equal(t, "timeout", res["status"])
}

func TestTrimOutput(t *testing.T) {
	big := strings.Repeat("a", codeOutputMaxBytes+10)
	out := trimOutput(big, "")
	assert.Contains(t, out, "截断")
	assert.LessOrEqual(t, len(out), codeOutputMaxBytes+50)
}
