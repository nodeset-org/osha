package eth_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/osha/vc/internal"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/node/validator/keymanager"
	"github.com/stretchr/testify/require"
)

func TestSetFeeRecipient(t *testing.T) {
	// Take a snapshot
	mgr.TakeSnapshot("test")
	defer func() {
		err := mgr.RevertToSnapshot("test")
		if err != nil {
			t.Fatalf("error reverting to snapshot: %v", err)
		}
	}()

	// Provision the database
	db := mgr.GetDatabase()
	internal.ProvisionDatabase(db)

	// Get the current fee recipients
	resp0 := runGetFeeRecipientRequest(t, internal.Pubkey0)
	resp1 := runGetFeeRecipientRequest(t, internal.Pubkey1)
	resp2 := runGetFeeRecipientRequest(t, internal.Pubkey2)

	// Make sure they're correct
	require.Equal(t, internal.Pubkey0, resp0.Pubkey)
	require.Equal(t, db.GetDefaultFeeRecipient(), resp0.EthAddress)
	require.Equal(t, internal.Pubkey1, resp1.Pubkey)
	require.Equal(t, internal.AltFeeRecipient, resp1.EthAddress)
	require.Equal(t, internal.Pubkey2, resp2.Pubkey)
	require.Equal(t, db.GetDefaultFeeRecipient(), resp2.EthAddress)
	t.Log("Received correct initial fee recipients")

	// Set the fee recipient for validator 2
	runPostFeeRecipientRequest(t, internal.Pubkey2, internal.AltFeeRecipient)
	t.Log("Set fee recipient for validator 2")

	// Get the new fee recipients
	fr0 := runGetFeeRecipientRequest(t, internal.Pubkey0).EthAddress
	fr1 := runGetFeeRecipientRequest(t, internal.Pubkey1).EthAddress
	fr2 := runGetFeeRecipientRequest(t, internal.Pubkey2).EthAddress

	// Make sure they're correct
	require.Equal(t, db.GetDefaultFeeRecipient(), fr0)
	require.Equal(t, internal.AltFeeRecipient, fr1)
	require.Equal(t, internal.AltFeeRecipient, fr2)
	t.Log("Received correct updated fee recipients")
}

// Run a GET eth/v1/validator/{pubkey}/feerecipient request
func runGetFeeRecipientRequest(t *testing.T, pubkey beacon.ValidatorPubkey) keymanager.GetFeeRecipientData {
	// Create the client
	client, err := keymanager.NewStandardKeyManagerClient(fmt.Sprintf("http://localhost:%d", port), jwtFilePath, nil)
	require.NoError(t, err)

	// Run the request
	data, err := client.GetFeeRecipientForValidator(context.Background(), logger, pubkey)
	require.NoError(t, err)
	t.Logf("Ran request")
	return data
}

// Run a POST eth/v1/validator/{pubkey}/feerecipient request
func runPostFeeRecipientRequest(t *testing.T, pubkey beacon.ValidatorPubkey, feeRecipient common.Address) {
	// Create the client
	client, err := keymanager.NewStandardKeyManagerClient(fmt.Sprintf("http://localhost:%d", port), jwtFilePath, nil)
	require.NoError(t, err)

	// Run the request
	err = client.SetFeeRecipientForValidator(context.Background(), logger, pubkey, feeRecipient)
	require.NoError(t, err)
	t.Logf("Ran request")
}
