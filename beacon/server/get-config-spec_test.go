package server

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/goccy/go-json"

	"github.com/nodeset-org/osha/beacon/api"

	"github.com/rocket-pool/node-manager-core/beacon/client"
	"github.com/rocket-pool/node-manager-core/utils"
	"github.com/stretchr/testify/require"
)

// Test getting the config spec
func TestConfigSpec(t *testing.T) {
	// Take a snapshot
	server.manager.TakeSnapshot("test")
	defer func() {
		err := server.manager.RevertToSnapshot("test")
		if err != nil {
			t.Fatalf("error reverting to snapshot: %v", err)
		}
	}()

	// Send a request
	parsedResponse := getConfigSpecResponse(t)

	// Make sure the response is correct
	cfg := server.manager.GetConfig()
	require.Equal(t, cfg.CapellaForkVersion, parsedResponse.Data.CapellaForkVersion)
	require.Equal(t, cfg.SecondsPerSlot, uint64(parsedResponse.Data.SecondsPerSlot))
	require.Equal(t, cfg.SlotsPerEpoch, uint64(parsedResponse.Data.SlotsPerEpoch))
	require.Equal(t, cfg.EpochsPerSyncCommitteePeriod, uint64(parsedResponse.Data.EpochsPerSyncCommitteePeriod))
	t.Logf(
		"Received correct response - seconds per slot: %d, slots per epoch: %d, epochs per sync committee: %d, capella fork: %s",
		uint64(parsedResponse.Data.SecondsPerSlot),
		uint64(parsedResponse.Data.SlotsPerEpoch),
		uint64(parsedResponse.Data.EpochsPerSyncCommitteePeriod),
		utils.EncodeHexWithPrefix(parsedResponse.Data.CapellaForkVersion),
	)
}

func getConfigSpecResponse(t *testing.T) client.Eth2ConfigResponse {
	// Create the request
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/eth/%s", port, api.ConfigSpecRoute), nil)
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
	var parsedResponse client.Eth2ConfigResponse
	err = json.Unmarshal(bytes, &parsedResponse)
	if err != nil {
		t.Fatalf("error deserializing response: %v", err)
	}

	t.Log("Parsed response")
	return parsedResponse
}
