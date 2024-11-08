package eth_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/nodeset-org/osha/vc/internal"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/node/validator/keymanager"
	"github.com/stretchr/testify/require"
)

func TestSetGraffiti(t *testing.T) {
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

	// Get the current graffiti
	resp0 := runGetGraffitiRequest(t, internal.Pubkey0)
	resp1 := runGetGraffitiRequest(t, internal.Pubkey1)
	resp2 := runGetGraffitiRequest(t, internal.Pubkey2)

	// Make sure they're correct
	require.Equal(t, internal.Pubkey0, resp0.Pubkey)
	require.Equal(t, db.GetDefaultGraffiti(), resp0.Graffiti)
	require.Equal(t, internal.Pubkey1, resp1.Pubkey)
	require.Equal(t, db.GetDefaultGraffiti(), resp1.Graffiti)
	require.Equal(t, internal.Pubkey2, resp2.Pubkey)
	require.Equal(t, internal.AltGraffiti, resp2.Graffiti)
	t.Log("Received correct initial graffiti")

	// Set the fee recipient for validator 1
	runPostGraffitiRequest(t, internal.Pubkey1, internal.AltGraffiti)
	t.Log("Set graffiti for validator 1")

	// Get the new fee recipients
	g0 := runGetGraffitiRequest(t, internal.Pubkey0).Graffiti
	g1 := runGetGraffitiRequest(t, internal.Pubkey1).Graffiti
	g2 := runGetGraffitiRequest(t, internal.Pubkey2).Graffiti

	// Make sure they're correct
	require.Equal(t, db.GetDefaultGraffiti(), g0)
	require.Equal(t, internal.AltGraffiti, g1)
	require.Equal(t, internal.AltGraffiti, g2)
	t.Log("Received correct updated graffiti")
}

// Run a GET eth/v1/validator/{pubkey}/graffiti request
func runGetGraffitiRequest(t *testing.T, pubkey beacon.ValidatorPubkey) keymanager.GetGraffitiData {
	// Create the client
	client, err := keymanager.NewStandardKeyManagerClient(fmt.Sprintf("http://localhost:%d", port), jwtFilePath, nil)
	require.NoError(t, err)

	// Run the request
	data, err := client.GetGraffitiForValidator(context.Background(), logger, pubkey)
	require.NoError(t, err)
	t.Logf("Ran request")
	return data
}

// Run a POST eth/v1/validator/{pubkey}/graffiti request
func runPostGraffitiRequest(t *testing.T, pubkey beacon.ValidatorPubkey, graffiti string) {
	// Create the client
	client, err := keymanager.NewStandardKeyManagerClient(fmt.Sprintf("http://localhost:%d", port), jwtFilePath, nil)
	require.NoError(t, err)

	// Run the request
	err = client.SetGraffitiForValidator(context.Background(), logger, pubkey, graffiti)
	require.NoError(t, err)
	t.Logf("Ran request")
}
