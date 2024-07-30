package server

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/goccy/go-json"

	"github.com/nodeset-org/osha/beacon/api"

	"github.com/rocket-pool/node-manager-core/beacon/client"
	"github.com/rocket-pool/node-manager-core/utils"
	"github.com/stretchr/testify/require"
)

// Test getting the Beacon genesis
func TestBeaconGenesis(t *testing.T) {
	// Take a snapshot
	server.manager.TakeSnapshot("test")
	defer func() {
		err := server.manager.RevertToSnapshot("test")
		if err != nil {
			t.Fatalf("error reverting to snapshot: %v", err)
		}
	}()

	// Send a request
	parsedResponse := getBeaconGenesisResponse(t)

	// Make sure the response is correct
	cfg := server.manager.GetConfig()
	genesisTime := time.Unix(int64(parsedResponse.Data.GenesisTime), 0)
	require.Equal(t, cfg.GenesisTime, genesisTime)
	require.Equal(t, cfg.GenesisValidatorsRoot, parsedResponse.Data.GenesisValidatorsRoot)
	require.Equal(t, cfg.GenesisForkVersion, parsedResponse.Data.GenesisForkVersion)
	t.Logf(
		"Received correct response - genesis time: %s, genesis root: %s, genesis fork: %s",
		genesisTime,
		utils.EncodeHexWithPrefix(parsedResponse.Data.GenesisValidatorsRoot),
		utils.EncodeHexWithPrefix(parsedResponse.Data.GenesisForkVersion),
	)
}

func getBeaconGenesisResponse(t *testing.T) client.GenesisResponse {
	// Create the request
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/eth/%s", port, api.BeaconGenesisRoute), nil)
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
	var parsedResponse client.GenesisResponse
	err = json.Unmarshal(bytes, &parsedResponse)
	if err != nil {
		t.Fatalf("error deserializing response: %v", err)
	}

	t.Log("Parsed response")
	return parsedResponse
}
