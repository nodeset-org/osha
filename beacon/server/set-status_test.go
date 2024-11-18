package server

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/nodeset-org/osha/beacon/api"
	idb "github.com/nodeset-org/osha/beacon/internal/db"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/stretchr/testify/require"
)

// Test setting a validator's status
func TestSetStatus(t *testing.T) {
	status := beacon.ValidatorState_ActiveOngoing

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
	v1 := d.GetValidatorByIndex(1)
	id := v1.Pubkey.HexWithPrefix()

	// Send the set status request
	sendSetStatusRequest(t, id, status)

	// Get the validator's status now
	parsedResponse := getValidatorResponse(t, id)

	// Make sure the response is correct
	require.Equal(t, string(status), parsedResponse.Data.Status)
	t.Logf("Received correct response - status: %s", parsedResponse.Data.Status)
}

func sendSetStatusRequest(t *testing.T, id string, status beacon.ValidatorState) {
	// Create the request
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/admin/%s", port, api.SetStatusRoute), nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}
	query := request.URL.Query()
	query.Add("id", id)
	query.Add("status", string(status))
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
