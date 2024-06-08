package server

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/nodeset-org/osha/beacon/api"
	idb "github.com/nodeset-org/osha/beacon/internal/db"
	"github.com/nodeset-org/osha/beacon/internal/test"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/node/validator"
	"github.com/stretchr/testify/require"
)

// Test setting a validator's balance
func TestAddValidator(t *testing.T) {
	pubkey, err := beacon.HexToValidatorPubkey(test.Pubkey3String)
	if err != nil {
		t.Fatalf("error converting pubkey [%s]: %v", test.Pubkey3String, err)
	}
	credsAddress := common.HexToAddress(test.WithdrawalCredentials2String)
	creds := validator.GetWithdrawalCredsFromAddress(credsAddress)

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

	// Send the set balance request
	parsedResponse := getAddValidatorResponse(t, pubkey, creds)
	require.Equal(t, uint64(3), parsedResponse.Index)
	t.Logf("Validator added with index %d", parsedResponse.Index)

	// Get the validator's status now
	id := strconv.FormatUint(parsedResponse.Index, 10)
	statusResponse := getValidatorResponse(t, id)

	// Make sure the response is correct
	require.Equal(t, pubkey, beacon.ValidatorPubkey(statusResponse.Data.Validator.Pubkey))
	require.Equal(t, creds, common.BytesToHash(statusResponse.Data.Validator.WithdrawalCredentials))
	t.Logf("Received correct response - pubkey: %s, creds: %s", pubkey.HexWithPrefix(), creds.Hex())
}

func getAddValidatorResponse(t *testing.T, pubkey beacon.ValidatorPubkey, creds common.Hash) api.AddValidatorResponse {
	// Create the request
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/admin/%s", port, api.AddValidatorRoute), nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}
	query := request.URL.Query()
	query.Add("pubkey", pubkey.HexWithPrefix())
	query.Add("creds", creds.Hex())
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
	var parsedResponse api.AddValidatorResponse
	err = json.Unmarshal(bytes, &parsedResponse)
	if err != nil {
		t.Fatalf("error deserializing response: %v", err)
	}

	t.Log("Parsed response")
	return parsedResponse
}
