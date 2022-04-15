package engine

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReachcounter(t *testing.T) {
	rc := newReachCounter()

	i := rc.Add("foo")
	require.Equal(t, uint(0), i)

	i = rc.Add("foo")
	require.Equal(t, uint(1), i)

	i = rc.Add("bar")
	require.Equal(t, uint(0), i)

	rc.Add("bar")
	i = rc.Add("bar")
	require.Equal(t, uint(2), i)

	rc.Clean()
	i = rc.Add("bar")
	require.Equal(t, uint(0), i)

	i = rc.Add("foo")
	require.Equal(t, uint(0), i)
}
