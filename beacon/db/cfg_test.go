package db

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigClone(t *testing.T) {
	c := NewDefaultConfig()
	clone := c.Clone()
	t.Log("Created config and clone")

	require.NotSame(t, c, clone)
	require.Equal(t, c, clone)
	t.Log("Configs are equal")
}
