package server

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/goccy/go-json"

	"github.com/nodeset-org/osha/beacon/api"
	idb "github.com/nodeset-org/osha/beacon/internal/db"
	"github.com/nodeset-org/osha/beacon/internal/test"

	"github.com/rocket-pool/node-manager-core/beacon/client"
	"github.com/stretchr/testify/require"
)

// Test getting the Beacon genesis
func TestBeaconHeaders(t *testing.T) {
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

	// Send a request
	response0 := getBeaconHeadersResponse(t, "0")
	response1 := getBeaconHeadersResponseWithRouteVar(t, "0")

	require.Equal(t, response0.Data.Root, test.BlockRootString)
	require.Equal(t, response0.Finalized, true)
	require.Equal(t, response1.Data.Root, test.BlockRootString)
	require.Equal(t, response1.Finalized, true)

	server.manager.CommitBlock(true)

	t.Logf(
		"Received correct response - data root slot 0: %s",
		response0.Data.Root,
	)
}

func getBeaconHeadersResponse(t *testing.T, slot string) client.BeaconBlockHeaderResponse {
	// Create the request
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/eth/%s", port, api.BeaconHeadersRoute), nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}

	query := request.URL.Query()
	query.Add("slot", slot)
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

	// Read the body
	defer response.Body.Close()
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("error reading the response body: %v", err)
	}
	var parsedResponse client.BeaconBlockHeaderResponse
	err = json.Unmarshal(bytes, &parsedResponse)
	if err != nil {
		t.Fatalf("error deserializing response: %v", err)
	}

	t.Log("Parsed response")
	return parsedResponse
}

func getBeaconHeadersResponseWithRouteVar(t *testing.T, slot string) client.BeaconBlockHeaderResponse {
	// Create the request
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/eth/%s", port, fmt.Sprintf(api.BeaconHeadersBlockIDRouteTemplate, slot)), nil)
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
	var parsedResponse client.BeaconBlockHeaderResponse
	err = json.Unmarshal(bytes, &parsedResponse)
	if err != nil {
		t.Fatalf("error deserializing response: %v", err)
	}

	t.Log("Parsed response")
	return parsedResponse
}
