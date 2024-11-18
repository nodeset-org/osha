package server

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/goccy/go-json"

	"github.com/nodeset-org/osha/beacon/api"
	"github.com/nodeset-org/osha/beacon/db"
	"github.com/rocket-pool/node-manager-core/beacon/client"
	"github.com/stretchr/testify/require"
)

// Test getting the deposit contract
func TestDepositContract(t *testing.T) {
	// Take a snapshot
	snapshotName, err := server.manager.TakeSnapshot()
	require.NoError(t, err)
	defer func() {
		err := server.manager.RevertToSnapshot(snapshotName)
		if err != nil {
			t.Fatalf("error reverting to snapshot: %v", err)
		}
	}()

	// Send a request
	parsedResponse := getDepositContractResponse(t)

	// Make sure the response is correct
	require.Equal(t, db.DefaultChainID, uint64(parsedResponse.Data.ChainID))
	require.Equal(t, db.DefaultDepositContractAddress, parsedResponse.Data.Address)
	t.Logf("Received correct response - chain ID: %d, deposit contract: %s", parsedResponse.Data.ChainID, parsedResponse.Data.Address.Hex())
}

func getDepositContractResponse(t *testing.T) client.Eth2DepositContractResponse {
	// Create the request
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/eth/%s", port, api.DepositContractRoute), nil)
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
	var parsedResponse client.Eth2DepositContractResponse
	err = json.Unmarshal(bytes, &parsedResponse)
	if err != nil {
		t.Fatalf("error deserializing response: %v", err)
	}

	t.Log("Parsed response")
	return parsedResponse
}
