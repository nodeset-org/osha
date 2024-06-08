package db

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatabaseClone(t *testing.T) {
	// Prep the pubkeys and creds
	logger := slog.Default()
	d := ProvisionDatabaseForTesting(t, logger)

	// Clone the database
	clone := d.Clone()
	require.NotSame(t, d, clone)
	require.Equal(t, d, clone)
	t.Log("Databases are equal")
}
