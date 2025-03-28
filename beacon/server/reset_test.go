package server

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/nodeset-org/osha/beacon/api"
	idb "github.com/nodeset-org/osha/beacon/internal/db"
	"github.com/stretchr/testify/require"
)

// Test slashing a validator
func TestReset(t *testing.T) {

	// Take a snapshot
	server.manager.TakeSnapshot("test")
	defer func() {
		err := server.manager.RevertToSnapshot("test")
		if err != nil {
			t.Fatalf("error reverting to snapshot: %v", err)
		}
	}()

	// Provision the database
	d := idb.ProvisionDatabaseForTesting(t, logger)

	server.manager.SetDatabase(d)

	response := getSyncStatusResponse(t)

	initialSlot := uint64(0)

	newSlot := uint64(1)

	require.Equal(t, initialSlot, uint64(response.Data.HeadSlot))

	sendCommitBlockRequest(t, true)

	response = getSyncStatusResponse(t)
	require.Equal(t, newSlot, uint64(response.Data.HeadSlot))

	// Send the slash request
	sendResetRequest(t)

	response = getSyncStatusResponse(t)
	require.Equal(t, initialSlot, uint64(response.Data.HeadSlot))

	t.Logf("Received correct response - head slot moved from %d to %d and then reset to %d", initialSlot, newSlot, initialSlot)
}

func sendResetRequest(t *testing.T) {
	// Create the request
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/admin/%s", port, api.ResetRoute), nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}
	t.Logf("Created request")

	// Send the request
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("error sending request: %v", err)
	}
	t.Logf("Sent request")

	// Check the status code
	require.Equal(t, http.StatusOK, response.StatusCode)
	t.Logf("Received OK status code")
}
