package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/goccy/go-json"
	"github.com/nodeset-org/osha/beacon/api"
	idb "github.com/nodeset-org/osha/beacon/internal/db"
	"github.com/rocket-pool/node-manager-core/beacon/client"
	"github.com/stretchr/testify/require"
)

// Test getting all validators
func TestAllValidators(t *testing.T) {
	// Take a snapshot
	snapshotName, err := server.manager.TakeSnapshot()
	require.NoError(t, err)
	defer func() {
		err := server.manager.RevertToSnapshot(snapshotName)
		if err != nil {
			t.Fatalf("error reverting to snapshot: %v", err)
		}
	}()

	// Provision the database
	d := idb.ProvisionDatabaseForTesting(t, logger)
	server.manager.SetDatabase(d)

	// Send a validator status request
	parsedResponse := getValidatorsResponse(t, nil)

	// Make sure the response is correct
	require.Len(t, parsedResponse.Data, 3)
	compareValidators(t, d.GetValidatorByIndex(0), &parsedResponse.Data[0])
	compareValidators(t, d.GetValidatorByIndex(1), &parsedResponse.Data[1])
	compareValidators(t, d.GetValidatorByIndex(2), &parsedResponse.Data[2])
	t.Log("Validators matched")
}

// Test getting 1 validator by index
func TestValidatorsByIndex_1(t *testing.T) {
	// Take a snapshot
	snapshotName, err := server.manager.TakeSnapshot()
	require.NoError(t, err)
	defer func() {
		err := server.manager.RevertToSnapshot(snapshotName)
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
	ids := []string{
		strconv.FormatUint(uint64(id), 10),
	}
	parsedResponse := getValidatorsResponse(t, ids)

	// Make sure the response is correct
	require.Len(t, parsedResponse.Data, 1)
	compareValidators(t, v, &parsedResponse.Data[0])
	t.Log("Validators matched")
}

// Test getting 1 validator by pubkey
func TestValidatorsByIndex_Pubkey(t *testing.T) {
	// Take a snapshot
	snapshotName, err := server.manager.TakeSnapshot()
	require.NoError(t, err)
	defer func() {
		err := server.manager.RevertToSnapshot(snapshotName)
		if err != nil {
			t.Fatalf("error reverting to snapshot: %v", err)
		}
	}()

	// Provision the database
	d := idb.ProvisionDatabaseForTesting(t, logger)
	server.manager.SetDatabase(d)

	// Send a validator status request
	v := d.GetValidatorByIndex(1)
	ids := []string{
		v.Pubkey.HexWithPrefix(),
	}
	parsedResponse := getValidatorsResponse(t, ids)

	// Make sure the response is correct
	require.Len(t, parsedResponse.Data, 1)
	compareValidators(t, v, &parsedResponse.Data[0])
	t.Log("Validators matched")
}

// Test getting multiple validators but not all
func TestValidatorsByIndex_Multi(t *testing.T) {
	// Take a snapshot
	snapshotName, err := server.manager.TakeSnapshot()
	require.NoError(t, err)
	defer func() {
		err := server.manager.RevertToSnapshot(snapshotName)
		if err != nil {
			t.Fatalf("error reverting to snapshot: %v", err)
		}
	}()

	// Provision the database
	d := idb.ProvisionDatabaseForTesting(t, logger)
	server.manager.SetDatabase(d)

	// Send a validator status request
	id0 := uint(0)
	v0 := d.GetValidatorByIndex(id0)
	v2 := d.GetValidatorByIndex(2)
	ids := []string{
		strconv.FormatUint(uint64(id0), 10),
		v2.Pubkey.HexWithPrefix(),
	}
	parsedResponse := getValidatorsResponse(t, ids)

	// Make sure the response is correct
	require.Len(t, parsedResponse.Data, 2)
	compareValidators(t, v0, &parsedResponse.Data[0])
	compareValidators(t, v2, &parsedResponse.Data[1])
	t.Log("Validators matched")
}

// Test getting multiple validators but not all via POST
func TestValidatorsByIndex_Post(t *testing.T) {
	// Take a snapshot
	snapshotName, err := server.manager.TakeSnapshot()
	require.NoError(t, err)
	defer func() {
		err := server.manager.RevertToSnapshot(snapshotName)
		if err != nil {
			t.Fatalf("error reverting to snapshot: %v", err)
		}
	}()

	// Provision the database
	d := idb.ProvisionDatabaseForTesting(t, logger)
	server.manager.SetDatabase(d)

	// Send a validator status request
	id0 := uint(0)
	v0 := d.GetValidatorByIndex(id0)
	v2 := d.GetValidatorByIndex(2)
	ids := []string{
		strconv.FormatUint(uint64(id0), 10),
		v2.Pubkey.HexWithPrefix(),
	}
	parsedResponse := getValidatorsResponsePost(t, ids)

	// Make sure the response is correct
	require.Len(t, parsedResponse.Data, 2)
	compareValidators(t, v0, &parsedResponse.Data[0])
	compareValidators(t, v2, &parsedResponse.Data[1])
	t.Log("Validators matched")
}

// Round trip a validators status request
func getValidatorsResponse(t *testing.T, ids []string) client.ValidatorsResponse {
	// Create the request
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/eth/%s", port, fmt.Sprintf(api.ValidatorsRouteTemplate, "head")), nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}
	query := request.URL.Query()
	query["id"] = ids
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
	var parsedResponse client.ValidatorsResponse
	err = json.Unmarshal(bytes, &parsedResponse)
	if err != nil {
		t.Fatalf("error deserializing response: %v", err)
	}

	t.Log("Parsed response")
	return parsedResponse
}

// Round trip a validators status request via a POST
func getValidatorsResponsePost(t *testing.T, ids []string) client.ValidatorsResponse {
	// Create the request
	reqBody := api.ValidatorsRequest{
		IDs: ids,
	}
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("error serializing request body: %v", err)
	}
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:%d/eth/%s", port, fmt.Sprintf(api.ValidatorsRouteTemplate, "head")), bytes.NewReader(reqBodyBytes))
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
	var parsedResponse client.ValidatorsResponse
	err = json.Unmarshal(bytes, &parsedResponse)
	if err != nil {
		t.Fatalf("error deserializing response: %v", err)
	}

	t.Log("Parsed response")
	return parsedResponse
}
