package team

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunKey(t *testing.T) {
	require.Equal(t, "cogniforge:team:run:u1:t1:default", runKey("u1", "t1", ""))
	require.Equal(t, "cogniforge:team:run:u1:t1:c9", runKey("u1", "t1", "c9"))
}

func TestNilStoreNoop(t *testing.T) {
	var s *RunStore
	require.NoError(t, s.Save(nil, "u", "t", "c", &RunState{TeamID: "t"}))
	got, err := s.Load(nil, "u", "t", "c")
	require.NoError(t, err)
	require.Nil(t, got)
}
