package server

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/nodeset-org/osha/beacon/api"
	idb "github.com/nodeset-org/osha/beacon/internal/db"
	"github.com/nodeset-org/osha/beacon/internal/test"
	"github.com/rocket-pool/node-manager-core/utils"
	"github.com/stretchr/testify/require"
)

// Test setting a slot's block root
func TestSetSlotExecutionBlockNumber(t *testing.T) {

	// Take a snapshot
	server.manager.TakeSnapshot("test")
	defer func() {
		err := server.manager.RevertToSnapshot("test")
		if err != nil {
			t.Fatalf("error reverting to snapshot: %v", err)
		}
	}()

	slotIndex := "0"
	newExecutionBlockNumber := uint64(100)

	// Provision the database
	d := idb.ProvisionDatabaseForTesting(t, logger)
	server.manager.SetDatabase(d)

	response := getBlindedBlocksResponse(t, slotIndex)

	require.Equal(t, response.Data.Message.Body.ExecutionPayloadHeader.BlockNumber, utils.Uinteger(test.ExecutionBlockNumber))

	// Send the set balance request
	getSetSlotExecutionBlockNumberResponse(t, 0, newExecutionBlockNumber)

	response = getBlindedBlocksResponse(t, slotIndex)

	require.Equal(t, response.Data.Message.Body.ExecutionPayloadHeader.BlockNumber, utils.Uinteger(newExecutionBlockNumber))

	t.Logf("Received correct response - slot: %s, execution block number changed from %d to %d", "0", test.ExecutionBlockNumber, newExecutionBlockNumber)

}

func getSetSlotExecutionBlockNumberResponse(t *testing.T, slot uint64, blockNumber uint64) {
	// Create the request
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/admin/%s", port, api.SetSlotExecutionBlockNumberRoute), nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}
	query := request.URL.Query()
	query.Add("slot", strconv.FormatUint(slot, 10))
	query.Add("block_number", strconv.FormatUint(blockNumber, 10))
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
