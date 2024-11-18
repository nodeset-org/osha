package server

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/nodeset-org/osha/beacon/api"
	idb "github.com/nodeset-org/osha/beacon/internal/db"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/stretchr/testify/require"
)

// Test slashing a validator
func TestSlash(t *testing.T) {
	penalty := uint64(1e9)

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

	// Make the validator active
	sendSetStatusRequest(t, id, beacon.ValidatorState_ActiveOngoing)
	t.Log("Marked the validator as active")

	// Get the original validator's status
	parsedResponse := getValidatorResponse(t, id)

	// Make sure the response is correct
	require.Equal(t, string(beacon.ValidatorState_ActiveOngoing), parsedResponse.Data.Status)
	require.Equal(t, uint64(32e9), uint64(parsedResponse.Data.Balance))
	require.False(t, parsedResponse.Data.Validator.Slashed)
	t.Logf("Original status is correct - status: %s", parsedResponse.Data.Status)

	// Send the slash request
	sendSlashRequest(t, id, penalty)

	// Get the validator's status now
	parsedResponse = getValidatorResponse(t, id)

	// Make sure the response is correct
	require.Equal(t, string(beacon.ValidatorState_ActiveSlashed), parsedResponse.Data.Status)
	require.Equal(t, uint64(32e9)-penalty, uint64(parsedResponse.Data.Balance))
	require.True(t, parsedResponse.Data.Validator.Slashed)
	t.Logf("Received correct response - status: %s, balance: %d, slashed: %t", parsedResponse.Data.Status, parsedResponse.Data.Balance, parsedResponse.Data.Validator.Slashed)
}

func sendSlashRequest(t *testing.T, id string, penalty uint64) {
	// Create the request
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/admin/%s", port, api.SlashRoute), nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}
	query := request.URL.Query()
	query.Add("id", id)
	query.Add("penalty", strconv.FormatUint(penalty, 10))
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
