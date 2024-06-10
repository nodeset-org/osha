package db

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/osha/beacon/internal/test"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/node/validator"
	"github.com/stretchr/testify/require"
)

func TestValidatorClone(t *testing.T) {
	pubkey, err := beacon.HexToValidatorPubkey(test.Pubkey0String)
	if err != nil {
		t.Fatalf("Error parsing pubkey: %v", err)
	}
	withdrawalCredsAddress := common.HexToAddress(test.WithdrawalCredentialsString)
	withdrawalCreds := validator.GetWithdrawalCredsFromAddress(withdrawalCredsAddress)
	v := NewValidator(pubkey, withdrawalCreds, 2)
	clone := v.Clone()
	t.Log("Created validator and clone")

	require.NotSame(t, v, clone)
	require.Equal(t, v, clone)
	t.Log("Validators are equal")
}
