package eth_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/nodeset-org/osha/vc/internal"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/node/validator/keymanager"
	"github.com/stretchr/testify/require"
)

func TestImportKeystore(t *testing.T) {
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

	// Get the current keystores
	keystores := runGetKeystoresRequest(t)
	require.Len(t, keystores, 3)
	var keystore0 keymanager.GetKeystoreData
	var keystore1 keymanager.GetKeystoreData
	var keystore2 keymanager.GetKeystoreData
	for _, keystore := range keystores {
		if keystore.Pubkey == internal.Pubkey0 {
			keystore0 = keystore
			continue
		}
		if keystore.Pubkey == internal.Pubkey1 {
			keystore1 = keystore
			continue
		}
		if keystore.Pubkey == internal.Pubkey2 {
			keystore2 = keystore
			continue
		}
	}
	require.Equal(t, keystore0.DerivationPath, "m/12381/3600/0/0/0")
	require.Equal(t, keystore1.DerivationPath, "m/12381/3600/1/0/0")
	require.Equal(t, keystore2.DerivationPath, "m/12381/3600/2/0/0")
	t.Log("Received correct initial keystores")

	// Add a new keystore
	pubkey3, _ := beacon.HexToValidatorPubkey("0x90de5e70beac090000000000000000000000000000000000000000000000000000000000000000000000000000000003")
	keystore := beacon.ValidatorKeystore{
		Crypto:  map[string]interface{}{},
		Name:    "validator3",
		Pubkey:  pubkey3,
		Version: 4,
		UUID:    uuid.MustParse("00000000-0000-0000-0000-000000000003"),
		Path:    "m/12381/3600/3/0/0",
	}
	password := "password3"
	data := runPostKeystoresRequest(t, []beacon.ValidatorKeystore{keystore}, []string{password}, nil)
	require.Len(t, data, 1)
	require.Equal(t, data[0].Status, keymanager.ImportKeystoreStatus_Imported)
	t.Log("Added new keystore")

	// Get the keystores again
	keystores = runGetKeystoresRequest(t)
	require.Len(t, keystores, 4)
	keystore0 = keymanager.GetKeystoreData{}
	keystore1 = keymanager.GetKeystoreData{}
	keystore2 = keymanager.GetKeystoreData{}
	var keystore3 keymanager.GetKeystoreData
	for _, keystore := range keystores {
		if keystore.Pubkey == internal.Pubkey0 {
			keystore0 = keystore
			continue
		}
		if keystore.Pubkey == internal.Pubkey1 {
			keystore1 = keystore
			continue
		}
		if keystore.Pubkey == internal.Pubkey2 {
			keystore2 = keystore
			continue
		}
		if keystore.Pubkey == pubkey3 {
			keystore3 = keystore
			continue
		}
	}
	require.Equal(t, keystore0.DerivationPath, "m/12381/3600/0/0/0")
	require.Equal(t, keystore1.DerivationPath, "m/12381/3600/1/0/0")
	require.Equal(t, keystore2.DerivationPath, "m/12381/3600/2/0/0")
	require.Equal(t, keystore3.DerivationPath, "m/12381/3600/3/0/0")
	t.Log("Received correct updated keystores")
}

func TestDeleteKeystore(t *testing.T) {
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

	// Get the current keystores
	keystores := runGetKeystoresRequest(t)
	require.Len(t, keystores, 3)
	var keystore0 keymanager.GetKeystoreData
	var keystore1 keymanager.GetKeystoreData
	var keystore2 keymanager.GetKeystoreData
	for _, keystore := range keystores {
		if keystore.Pubkey == internal.Pubkey0 {
			keystore0 = keystore
			continue
		}
		if keystore.Pubkey == internal.Pubkey1 {
			keystore1 = keystore
			continue
		}
		if keystore.Pubkey == internal.Pubkey2 {
			keystore2 = keystore
			continue
		}
	}
	require.Equal(t, keystore0.DerivationPath, "m/12381/3600/0/0/0")
	require.Equal(t, keystore1.DerivationPath, "m/12381/3600/1/0/0")
	require.Equal(t, keystore2.DerivationPath, "m/12381/3600/2/0/0")
	t.Log("Received correct initial keystores")

	// Delete a keystore
	data, slashingProtection := runDeleteKeystoresRequest(t, []beacon.ValidatorPubkey{internal.Pubkey2})
	require.Len(t, data, 1)
	require.Equal(t, data[0].Status, keymanager.DeleteKeystoreStatus_Deleted)
	require.Len(t, slashingProtection.Data, 1)
	require.Equal(t, slashingProtection.Data[0].Pubkey, internal.Pubkey2)
	t.Log("Deleted keystore")

	// Get the keystores again
	keystores = runGetKeystoresRequest(t)
	require.Len(t, keystores, 2)
	keystore0 = keymanager.GetKeystoreData{}
	keystore1 = keymanager.GetKeystoreData{}
	for _, keystore := range keystores {
		if keystore.Pubkey == internal.Pubkey0 {
			keystore0 = keystore
			continue
		}
		if keystore.Pubkey == internal.Pubkey1 {
			keystore1 = keystore
			continue
		}
	}
	require.Equal(t, keystore0.DerivationPath, "m/12381/3600/0/0/0")
	require.Equal(t, keystore1.DerivationPath, "m/12381/3600/1/0/0")
	t.Log("Received correct initial keystores")
}

// Run a GET eth/v1/keystores request
func runGetKeystoresRequest(t *testing.T) []keymanager.GetKeystoreData {
	// Create the client
	client, err := keymanager.NewStandardKeyManagerClient(fmt.Sprintf("http://localhost:%d", port), jwtFilePath, nil)
	require.NoError(t, err)

	// Run the request
	data, err := client.GetLoadedKeys(context.Background(), logger)
	require.NoError(t, err)
	t.Logf("Ran request")
	return data
}

// Run a POST eth/v1/keystores request
func runPostKeystoresRequest(t *testing.T, keystores []beacon.ValidatorKeystore, passwords []string, slashingProtection *beacon.SlashingProtectionData) []keymanager.ImportKeystoreData {
	// Create the client
	client, err := keymanager.NewStandardKeyManagerClient(fmt.Sprintf("http://localhost:%d", port), jwtFilePath, nil)
	require.NoError(t, err)

	// Run the request
	data, err := client.ImportKeys(context.Background(), logger, keystores, passwords, slashingProtection)
	require.NoError(t, err)
	t.Logf("Ran request")
	return data
}

// Run a DELETE eth/v1/keystores request
func runDeleteKeystoresRequest(t *testing.T, pubkeys []beacon.ValidatorPubkey) ([]keymanager.DeleteKeystoreData, *beacon.SlashingProtectionData) {
	// Create the client
	client, err := keymanager.NewStandardKeyManagerClient(fmt.Sprintf("http://localhost:%d", port), jwtFilePath, nil)
	require.NoError(t, err)

	// Run the request
	data, slashingProtection, err := client.DeleteKeys(context.Background(), logger, pubkeys)
	require.NoError(t, err)
	t.Logf("Ran request")
	return data, slashingProtection
}
