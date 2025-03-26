package db

import (
	"log/slog"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/osha/beacon/db"
	"github.com/nodeset-org/osha/beacon/internal/test"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/node/validator"
	"github.com/stretchr/testify/require"
)

func ProvisionDatabaseForTesting(t *testing.T, logger *slog.Logger) *db.Database {
	// Prep the pubkeys and creds
	pubkey0, err := beacon.HexToValidatorPubkey(test.Pubkey0String)
	if err != nil {
		t.Fatalf("Error parsing pubkey [%s]: %v", test.Pubkey0String, err)
	}
	pubkey1, err := beacon.HexToValidatorPubkey(test.Pubkey1String)
	if err != nil {
		t.Fatalf("Error parsing pubkey [%s]: %v", test.Pubkey1String, err)
	}
	pubkey2, err := beacon.HexToValidatorPubkey(test.Pubkey2String)
	if err != nil {
		t.Fatalf("Error parsing pubkey [%s]: %v", test.Pubkey2String, err)
	}
	withdrawalCredsAddress := common.HexToAddress(test.WithdrawalCredentialsString)
	withdrawalCreds := validator.GetWithdrawalCredsFromAddress(withdrawalCredsAddress)
	t.Log("Prepped pubkeys and creds")

	// Create a new database
	d := db.NewDatabase(logger, 0)
	v0, err := d.AddValidator(pubkey0, withdrawalCreds)
	if err != nil {
		t.Fatalf("Error adding validator [%s]: %v", pubkey0.HexWithPrefix(), err)
	}
	v1, err := d.AddValidator(pubkey1, withdrawalCreds)
	if err != nil {
		t.Fatalf("Error adding validator [%s]: %v", pubkey1.HexWithPrefix(), err)
	}
	v2, err := d.AddValidator(pubkey2, withdrawalCreds)
	if err != nil {
		t.Fatalf("Error adding validator [%s]: %v", pubkey1.HexWithPrefix(), err)
	}

	require.Same(t, v0, d.GetValidatorByIndex(0))
	require.Same(t, v1, d.GetValidatorByIndex(1))
	require.Same(t, v2, d.GetValidatorByIndex(2))
	require.NotSame(t, v0, v1)
	require.NotSame(t, v1, v2)
	t.Log("Added validators to database")

	b1, _ := d.SetSlotBlockRoot(1, common.HexToHash(test.BlockRootString))

	require.Equal(t, b1, true)
	t.Log("Set block root for slot 1")

	return d
}
