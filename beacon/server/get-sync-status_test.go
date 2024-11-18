package server

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/goccy/go-json"

	"github.com/nodeset-org/osha/beacon/api"
	idb "github.com/nodeset-org/osha/beacon/internal/db"
	"github.com/rocket-pool/node-manager-core/beacon/client"
	"github.com/stretchr/testify/require"
)

// Make sure sync status requests work when synced
func TestSynced(t *testing.T) {
	currentSlot := uint64(12)

	// Take a snapshot
	snapshotName, err := server.manager.TakeSnapshot()
	require.NoError(t, err)
	defer func() {
		err := server.manager.RevertToSnapshot(snapshotName)
		if err != nil {
			t.Fatalf("error reverting to snapshot: %v", err)
		}
	}()

	// Provision the database
	d := idb.ProvisionDatabaseForTesting(t, logger)
	server.manager.SetDatabase(d)
	for i := uint64(0); i < currentSlot; i++ {
		server.manager.CommitBlock(true)
	}

	// Send a sync status request
	parsedResponse := getSyncStatusResponse(t)

	// Make sure the response is correct
	require.Equal(t, currentSlot, uint64(parsedResponse.Data.HeadSlot))
	require.Equal(t, uint64(0), uint64(parsedResponse.Data.SyncDistance))
	require.False(t, parsedResponse.Data.IsSyncing)
	t.Logf("Received correct response - head slot: %d, sync distance: %d, is syncing: %t", parsedResponse.Data.HeadSlot, parsedResponse.Data.SyncDistance, parsedResponse.Data.IsSyncing)
}

// Make sure sync status requests work when still syncing
func TestUnsynced(t *testing.T) {
	currentSlot := uint64(8)
	headSlot := uint64(12)

	// Take a snapshot
	snapshotName, err := server.manager.TakeSnapshot()
	require.NoError(t, err)
	defer func() {
		err := server.manager.RevertToSnapshot(snapshotName)
		if err != nil {
			t.Fatalf("error reverting to snapshot: %v", err)
		}
	}()

	// Provision the database
	d := idb.ProvisionDatabaseForTesting(t, logger)
	server.manager.SetDatabase(d)
	server.manager.SetHighestSlot(headSlot)
	for i := uint64(0); i < currentSlot; i++ {
		server.manager.CommitBlock(true)
	}

	// Send a sync status request
	parsedResponse := getSyncStatusResponse(t)

	// Make sure the response is correct
	require.Equal(t, headSlot, uint64(parsedResponse.Data.HeadSlot))
	require.Equal(t, headSlot-currentSlot, uint64(parsedResponse.Data.SyncDistance))
	require.True(t, parsedResponse.Data.IsSyncing)
	t.Logf("Received correct response - head slot: %d, sync distance: %d, is syncing: %t", parsedResponse.Data.HeadSlot, parsedResponse.Data.SyncDistance, parsedResponse.Data.IsSyncing)
}

func getSyncStatusResponse(t *testing.T) client.SyncStatusResponse {
	// Create the request
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/eth/%s", port, api.SyncingRoute), nil)
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

	// Read the body
	defer response.Body.Close()
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("error reading the response body: %v", err)
	}
	var parsedResponse client.SyncStatusResponse
	err = json.Unmarshal(bytes, &parsedResponse)
	if err != nil {
		t.Fatalf("error deserializing response: %v", err)
	}

	t.Log("Parsed response")
	return parsedResponse
}
