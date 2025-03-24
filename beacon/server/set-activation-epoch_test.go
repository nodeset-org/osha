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

// Test setting a validator's status
func TestSetActivationEpoch(t *testing.T) {
	activationEpoch := uint64(1100)

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
	v1 := d.GetValidatorByIndex(1)
	id := v1.Pubkey.HexWithPrefix()

	// Send the set status request
	sendSetActivationEpochRequest(t, id, activationEpoch)

	// Get the validator's status now
	parsedResponse := getValidatorResponse(t, id)

	// Make sure the response is correct
	require.Equal(t, activationEpoch, uint64(parsedResponse.Data.Validator.ActivationEpoch))
	t.Logf("Received correct response - status: %d", parsedResponse.Data.Validator.ActivationEpoch)
}

func sendSetActivationEpochRequest(t *testing.T, id string, epoch uint64) {
	// Create the request
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/admin/%s", port, api.SetActivationEpochRoute), nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}
	query := request.URL.Query()
	query.Add("id", id)
	query.Add("epoch", strconv.FormatUint(epoch, 10))
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
