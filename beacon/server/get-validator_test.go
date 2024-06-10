package server

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/goccy/go-json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/osha/beacon/api"
	"github.com/nodeset-org/osha/beacon/db"
	idb "github.com/nodeset-org/osha/beacon/internal/db"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/beacon/client"
	"github.com/stretchr/testify/require"
)

// Test getting a validator by index
func TestValidatorIndex(t *testing.T) {
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

	// Send a validator status request
	id := uint(1)
	v := d.GetValidatorByIndex(id)
	require.NotNil(t, v)
	parsedResponse := getValidatorResponse(t, strconv.FormatUint(uint64(id), 10))

	// Make sure the response is correct
	compareValidators(t, v, &parsedResponse.Data)
	t.Log("Validators matched")
}

// Test getting a validator by pubkey
func TestValidatorPubkey(t *testing.T) {
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

	// Send a validator status request
	v := d.GetValidatorByIndex(2)
	require.NotNil(t, v)
	parsedResponse := getValidatorResponse(t, v.Pubkey.HexWithPrefix())

	// Make sure the response is correct
	compareValidators(t, v, &parsedResponse.Data)
	t.Log("Validators matched")
}

// Round trip a validator status request
func getValidatorResponse(t *testing.T, id string) api.ValidatorResponse {
	// Create the request
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/eth/%s", port, fmt.Sprintf(api.ValidatorRouteTemplate, "head", id)), nil)
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
	var parsedResponse api.ValidatorResponse
	err = json.Unmarshal(bytes, &parsedResponse)
	if err != nil {
		t.Fatalf("error deserializing response: %v", err)
	}

	t.Log("Parsed response")
	return parsedResponse
}

func compareValidators(t *testing.T, local *db.Validator, remote *client.Validator) {
	require.Equal(t, strconv.FormatUint(local.Index, 10), remote.Index)
	require.Equal(t, local.Balance, uint64(remote.Balance))
	require.Equal(t, string(local.Status), remote.Status)
	require.Equal(t, local.Pubkey, beacon.ValidatorPubkey(remote.Validator.Pubkey))
	require.Equal(t, local.WithdrawalCredentials, common.BytesToHash(remote.Validator.WithdrawalCredentials))
	require.Equal(t, local.EffectiveBalance, uint64(remote.Validator.EffectiveBalance))
	require.Equal(t, local.Slashed, remote.Validator.Slashed)
	require.Equal(t, local.ActivationEligibilityEpoch, uint64(remote.Validator.ActivationEligibilityEpoch))
	require.Equal(t, local.ActivationEpoch, uint64(remote.Validator.ActivationEpoch))
	require.Equal(t, local.ExitEpoch, uint64(remote.Validator.ExitEpoch))
	require.Equal(t, local.WithdrawableEpoch, uint64(remote.Validator.WithdrawableEpoch))
}
