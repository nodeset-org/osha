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

// Test setting a validator's balance
func TestSetBalance(t *testing.T) {
	balance := uint64(33e9)

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

	// Send the set balance request
	sendSetBalanceRequest(t, id, balance)

	// Get the validator's status now
	parsedResponse := getValidatorResponse(t, id)

	// Make sure the response is correct
	require.Equal(t, balance, uint64(parsedResponse.Data.Balance))
	t.Logf("Received correct response - balance: %d", parsedResponse.Data.Balance)
}

func sendSetBalanceRequest(t *testing.T, id string, balance uint64) {
	// Create the request
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/admin/%s", port, api.SetBalanceRoute), nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}
	query := request.URL.Query()
	query.Add("id", id)
	query.Add("balance", strconv.FormatUint(balance, 10))
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
