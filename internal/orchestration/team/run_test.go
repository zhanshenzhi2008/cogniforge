package team

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunTaskRequiresTask(t *testing.T) {
	deps := &ChatDeps{}
	_, err := deps.RunTask(nil, "u", nil, "  ", "", nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "task")
}
