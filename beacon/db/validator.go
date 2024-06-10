package db

import (
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/beacon/client"
)

type Validator struct {
	Pubkey                     beacon.ValidatorPubkey
	Index                      uint64
	WithdrawalCredentials      common.Hash
	Balance                    uint64
	Status                     beacon.ValidatorState
	EffectiveBalance           uint64
	Slashed                    bool
	ActivationEligibilityEpoch uint64
	ActivationEpoch            uint64
	ExitEpoch                  uint64
	WithdrawableEpoch          uint64
}

func NewValidator(pubkey beacon.ValidatorPubkey, withdrawalCredentials common.Hash, index uint64) *Validator {
	return &Validator{
		Pubkey:                     pubkey,
		Index:                      index,
		WithdrawalCredentials:      withdrawalCredentials,
		Balance:                    StartingBalance,
		Status:                     beacon.ValidatorState_PendingInitialized,
		EffectiveBalance:           StartingBalance,
		Slashed:                    false,
		ActivationEligibilityEpoch: FarFutureEpoch,
		ActivationEpoch:            FarFutureEpoch,
		ExitEpoch:                  FarFutureEpoch,
		WithdrawableEpoch:          FarFutureEpoch,
	}
}

func (v *Validator) SetBalance(balanceGwei uint64) {
	v.Balance = balanceGwei

	// Rules taken from the spec: https://github.com/ethereum/annotated-spec/blob/master/phase0/beacon-chain.md#misc
	if balanceGwei < v.EffectiveBalance-25e7 {
		v.EffectiveBalance--
	}
	if balanceGwei > v.EffectiveBalance+125e7 {
		v.EffectiveBalance++
	}
}

func (v *Validator) SetStatus(status beacon.ValidatorState) {
	v.Status = status
}

func (v *Validator) Slash(penaltyGwei uint64) error {
	if v.Status != beacon.ValidatorState_ActiveOngoing && v.Status != beacon.ValidatorState_ActiveExiting {
		return fmt.Errorf("validator with pubkey %s is not in a slashable state", v.Pubkey.HexWithPrefix())
	}
	v.Slashed = true
	v.SetBalance(v.Balance - penaltyGwei)
	v.Status = beacon.ValidatorState_ActiveSlashed
	return nil
}

func (v *Validator) GetValidatorMeta() client.Validator {
	validatorMeta := client.Validator{
		Index:   strconv.FormatUint(v.Index, 10),
		Balance: client.Uinteger(v.Balance),
		Status:  string(v.Status),
	}
	validatorMeta.Validator.Pubkey = v.Pubkey[:]
	validatorMeta.Validator.WithdrawalCredentials = v.WithdrawalCredentials[:]
	validatorMeta.Validator.EffectiveBalance = client.Uinteger(v.EffectiveBalance)
	validatorMeta.Validator.Slashed = v.Slashed
	validatorMeta.Validator.ActivationEligibilityEpoch = client.Uinteger(v.ActivationEligibilityEpoch)
	validatorMeta.Validator.ActivationEpoch = client.Uinteger(v.ActivationEpoch)
	validatorMeta.Validator.ExitEpoch = client.Uinteger(v.ExitEpoch)
	validatorMeta.Validator.WithdrawableEpoch = client.Uinteger(v.WithdrawableEpoch)
	return validatorMeta
}

func (v *Validator) Clone() *Validator {
	return &Validator{
		Pubkey:                     v.Pubkey,
		Index:                      v.Index,
		WithdrawalCredentials:      v.WithdrawalCredentials,
		Balance:                    v.Balance,
		Status:                     v.Status,
		EffectiveBalance:           v.EffectiveBalance,
		Slashed:                    v.Slashed,
		ActivationEligibilityEpoch: v.ActivationEligibilityEpoch,
		ActivationEpoch:            v.ActivationEpoch,
		ExitEpoch:                  v.ExitEpoch,
		WithdrawableEpoch:          v.WithdrawableEpoch,
	}
}
