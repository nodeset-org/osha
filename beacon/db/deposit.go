package db

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/beacon/client"
	"github.com/rocket-pool/node-manager-core/utils"
)

type Deposit struct {
	// The validator's public key
	Pubkey beacon.ValidatorPubkey

	// The validator's withdrawal credentials
	WithdrawalCredentials common.Hash

	// The amount of ETH deposited, in gwei
	Amount uint64

	// The deposit signature
	Signature beacon.ValidatorSignature

	// The slot that the deposit was made in
	Slot uint64
}

func (d Deposit) ConvertToNativeFormat() client.PendingDeposit {
	return client.PendingDeposit{
		Pubkey:                d.Pubkey[:],
		WithdrawalCredentials: d.WithdrawalCredentials[:],
		Amount:                utils.Uinteger(d.Amount),
		Signature:             d.Signature[:],
		Slot:                  utils.Uinteger(d.Slot),
	}
}
