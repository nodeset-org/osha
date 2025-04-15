package server

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/goccy/go-json"
	"github.com/nodeset-org/osha/beacon/api"
	"github.com/nodeset-org/osha/beacon/db"
	idb "github.com/nodeset-org/osha/beacon/internal/db"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/beacon/client"
	"github.com/stretchr/testify/require"
)

// Test getting pending deposits
func TestPendingDeposits(t *testing.T) {
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

	// Add some pending deposit
	validator0 := d.GetValidatorByIndex(0)
	validator1 := d.GetValidatorByIndex(1)
	pendingDeposit := db.Deposit{
		Pubkey:                validator0.Pubkey,
		WithdrawalCredentials: validator0.WithdrawalCredentials,
		Amount:                32 * 1e9, // 32 ETH in gwei
		Signature:             beacon.ValidatorSignature{0x01},
		Slot:                  0,
	}
	d.AddPendingDeposit(&pendingDeposit)
	pendingDeposit2 := db.Deposit{
		Pubkey:                validator1.Pubkey,
		WithdrawalCredentials: validator1.WithdrawalCredentials,
		Amount:                32 * 1e9, // 32 ETH in gwei
		Signature:             beacon.ValidatorSignature{0x02},
		Slot:                  d.GetHighestSlot(),
	}
	d.AddPendingDeposit(&pendingDeposit2)

	// Send a request
	v := d.GetPendingDeposits()
	require.NotNil(t, v)
	parsedResponse := getPendingDepositsResponse(t)

	// Make sure the response is correct
	for i, deposit := range parsedResponse.Data {
		local := v[i]
		comparePendingDeposit(t, local, deposit)
	}
	t.Log("Pending deposits matched")
}

// Round trip a validator pending deposits request
func getPendingDepositsResponse(t *testing.T) client.PendingDepositsResponse {
	// Create the request
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/eth/%s", port, fmt.Sprintf(api.PendingDepositsRouteTemplate, "head")), nil)
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
	var parsedResponse client.PendingDepositsResponse
	err = json.Unmarshal(bytes, &parsedResponse)
	if err != nil {
		t.Fatalf("error deserializing response: %v", err)
	}

	t.Log("Parsed response")
	return parsedResponse
}

func comparePendingDeposit(t *testing.T, local *db.Deposit, remote client.PendingDeposit) {
	native := local.ConvertToNativeFormat()
	require.Equal(t, native.Pubkey, remote.Pubkey)
	require.Equal(t, native.WithdrawalCredentials, remote.WithdrawalCredentials)
	require.Equal(t, native.Amount, remote.Amount)
	require.Equal(t, native.Signature, remote.Signature)
	require.Equal(t, native.Slot, remote.Slot)
}
