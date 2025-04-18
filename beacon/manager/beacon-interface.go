package manager

import (
	"context"

	"github.com/rocket-pool/node-manager-core/beacon/client"
	"github.com/rocket-pool/node-manager-core/utils"
)

// Temp until finality is implemented
func (m *BeaconMockManager) Beacon_FinalityCheckpoints(ctx context.Context, stateId string) (client.FinalityCheckpointsResponse, error) {
	response := client.FinalityCheckpointsResponse{}
	response.Data.Finalized.Epoch = utils.Uinteger(m.database.GetCurrentSlot())
	response.Data.CurrentJustified.Epoch = utils.Uinteger(m.database.GetCurrentSlot())
	response.Data.PreviousJustified.Epoch = utils.Uinteger(m.database.GetCurrentSlot())
	return response, nil
}

func (m *BeaconMockManager) Beacon_Genesis(ctx context.Context) (client.GenesisResponse, error) {
	response := client.GenesisResponse{}
	response.Data.GenesisTime = utils.Uinteger(m.config.GenesisTime.Unix())
	response.Data.GenesisValidatorsRoot = m.config.GenesisValidatorsRoot
	response.Data.GenesisForkVersion = m.config.GenesisForkVersion
	return response, nil
}

func (m *BeaconMockManager) Beacon_Validators(ctx context.Context, stateId string, ids []string) (client.ValidatorsResponse, error) {
	// Get the validators
	validators, err := m.GetValidators(ids)
	if err != nil {
		return client.ValidatorsResponse{}, err
	}

	// Write the response
	validatorMetas := make([]client.Validator, len(validators))
	for i, validator := range validators {
		validatorMetas[i] = validator.GetValidatorMeta()
	}
	response := client.ValidatorsResponse{
		Data: validatorMetas,
	}
	return response, nil
}

func (m *BeaconMockManager) Beacon_PendingDeposits(ctx context.Context, stateID string) (client.PendingDepositsResponse, error) {
	// Get the pending deposits
	pendingDeposits := m.GetPendingDeposits()

	// Convert the deposit data to the native format
	nativeDeposits := make([]client.PendingDeposit, len(pendingDeposits))
	for i, deposit := range pendingDeposits {
		nativeDeposits[i] = deposit.ConvertToNativeFormat()
	}

	// Write the response
	response := client.PendingDepositsResponse{
		Data: nativeDeposits,
	}
	return response, nil
}

func (m *BeaconMockManager) Config_DepositContract(ctx context.Context) (client.Eth2DepositContractResponse, error) {
	response := client.Eth2DepositContractResponse{}
	response.Data.Address = m.config.DepositContract
	response.Data.ChainID = utils.Uinteger(m.config.ChainID)
	return response, nil
}

func (m *BeaconMockManager) Config_Spec(ctx context.Context) (client.Eth2ConfigResponse, error) {
	response := client.Eth2ConfigResponse{}
	response.Data.SecondsPerSlot = utils.Uinteger(m.config.SecondsPerSlot)
	response.Data.SlotsPerEpoch = utils.Uinteger(m.config.SlotsPerEpoch)
	response.Data.EpochsPerSyncCommitteePeriod = utils.Uinteger(m.config.EpochsPerSyncCommitteePeriod)
	response.Data.CapellaForkVersion = m.config.CapellaForkVersion
	return response, nil
}

func (m *BeaconMockManager) Node_Syncing(ctx context.Context) (client.SyncStatusResponse, error) {
	// Get the slots
	currentSlot := m.GetCurrentSlot()
	highestSlot := m.GetHighestSlot()

	// Write the response
	response := client.SyncStatusResponse{}
	response.Data.IsSyncing = (currentSlot < highestSlot)
	response.Data.HeadSlot = utils.Uinteger(highestSlot)
	response.Data.SyncDistance = utils.Uinteger(highestSlot - currentSlot)
	return response, nil
}
