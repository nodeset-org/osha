package server

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/rocket-pool/node-manager-core/utils"

	"github.com/nodeset-org/osha/beacon/api"
	idb "github.com/nodeset-org/osha/beacon/internal/db"
	"github.com/nodeset-org/osha/beacon/internal/test"

	"github.com/stretchr/testify/require"
)

// Test getting the Beacon genesis
func TestBlindedBlocks(t *testing.T) {
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

	slotIndex := "0"
	executionBlockNumber := test.ExecutionBlockNumber

	newBlockRoot := common.HexToHash("0x1234")
	newExecutionBlockNumber := uint64(100)

	// Send a request
	response := getBlindedBlocksResponse(t, slotIndex)

	slotIndexUint, err := strconv.ParseUint(slotIndex, 10, 64)

	if err != nil {
		t.Fatalf("error parsing slot index [%s]: %v", slotIndex, err)
	}

	require.Equal(t, response.Data.Message.Slot, utils.Uinteger(slotIndexUint))
	require.Equal(t, response.Data.Message.Body.ExecutionPayloadHeader.BlockNumber, utils.Uinteger(test.ExecutionBlockNumber))

	t.Logf(
		"Received correct response - initial execution block number: %d",
		executionBlockNumber,
	)

	server.manager.SetSlotBlockRoot(slotIndexUint, newBlockRoot)
	server.manager.SetSlotExecutionBlockNumber(slotIndexUint, newExecutionBlockNumber)

	newResponseBySlotIndex := getBlindedBlocksResponse(t, slotIndex)
	newResponseByBlockRoot := getBlindedBlocksResponse(t, newBlockRoot.Hex())

	require.Equal(t, newResponseBySlotIndex.Data.Message.Body.ExecutionPayloadHeader.BlockNumber, utils.Uinteger(newExecutionBlockNumber))
	require.Equal(t, newResponseBySlotIndex, newResponseByBlockRoot)

	t.Logf(
		"Received correct response - slot: %s, execution block number changed from %d to %d",
		slotIndex,
		executionBlockNumber,
		newExecutionBlockNumber,
	)

	t.Logf(
		"Received correct response - slot accessible by slot index: %s and new block root: %s",
		slotIndex,
		newBlockRoot.Hex(),
	)

}

func getBlindedBlocksResponse(t *testing.T, block_id string) api.BlindedBlockResponse {
	// Create the request
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/eth/%s", port, fmt.Sprintf(api.BlindedBlocksRouteTemplate, block_id)), nil)
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
	var parsedResponse api.BlindedBlockResponse
	err = json.Unmarshal(bytes, &parsedResponse)
	if err != nil {
		t.Fatalf("error deserializing response: %v", err)
	}

	t.Log("Parsed response")
	return parsedResponse
}
