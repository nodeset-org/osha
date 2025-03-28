package server

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/osha/beacon/api"
	idb "github.com/nodeset-org/osha/beacon/internal/db"
	"github.com/nodeset-org/osha/beacon/internal/test"
	"github.com/stretchr/testify/require"
)

// Test setting a slot's block root
func TestSetSlotBlockRoot(t *testing.T) {

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

	response := getBeaconHeadersResponse(t, "0")

	require.Equal(t, response.Data.Root, test.BlockRootString)

	newSlotBlockRoot := common.HexToHash("0x2234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")

	// Send the set balance request
	getSetSlotBlockRootResponse(t, 0, newSlotBlockRoot)

	response = getBeaconHeadersResponse(t, "0")

	require.Equal(t, response.Data.Root, newSlotBlockRoot.String())
	t.Logf("Received correct response - slot: %s block root transition from %s to %s", "0", test.BlockRootString, newSlotBlockRoot.String())

}

func getSetSlotBlockRootResponse(t *testing.T, slot uint64, root common.Hash) {
	// Create the request
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/admin/%s", port, api.SetSlotBlockRootRoute), nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}
	query := request.URL.Query()
	query.Add("slot", strconv.FormatUint(slot, 10))
	query.Add("root", root.Hex())
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
