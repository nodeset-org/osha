package server

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/nodeset-org/osha/beacon/api"
	idb "github.com/nodeset-org/osha/beacon/internal/db"
	"github.com/stretchr/testify/require"
)

// Test committing a block to the chain
func TestCommitBlock(t *testing.T) {
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

	// Send a commit request
	sendCommitBlockRequest(t, true)

	// Get the sync status now
	parsedResponse := getSyncStatusResponse(t)

	// Make sure the response is correct
	require.Equal(t, uint64(1), uint64(parsedResponse.Data.HeadSlot))
	require.Equal(t, uint64(0), uint64(parsedResponse.Data.SyncDistance))
	require.False(t, parsedResponse.Data.IsSyncing)
	t.Logf("Received correct response - head slot: %d, sync distance: %d, is syncing: %t", parsedResponse.Data.HeadSlot, parsedResponse.Data.SyncDistance, parsedResponse.Data.IsSyncing)
}

func sendCommitBlockRequest(t *testing.T, slotValidated bool) {
	// Create the request
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/admin/%s", port, api.CommitBlockRoute), nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}
	query := request.URL.Query()
	query.Add("validated", strconv.FormatBool(slotValidated))
	request.URL.RawQuery = query.Encode()
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
