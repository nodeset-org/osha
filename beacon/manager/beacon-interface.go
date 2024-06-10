package manager

import (
	"context"
	"fmt"

	"github.com/rocket-pool/node-manager-core/beacon/client"
)

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

func (m *BeaconMockManager) Config_DepositContract(ctx context.Context) (client.Eth2DepositContractResponse, error) {
	response := client.Eth2DepositContractResponse{}
	response.Data.Address = m.config.DepositContract
	response.Data.ChainID = client.Uinteger(m.config.ChainID)
	return response, nil
}

func (m *BeaconMockManager) Node_Syncing(ctx context.Context) (client.SyncStatusResponse, error) {
	// Get the slots
	currentSlot := m.GetCurrentSlot()
	highestSlot := m.GetHighestSlot()

	// Write the response
	response := client.SyncStatusResponse{}
	response.Data.IsSyncing = (currentSlot < highestSlot)
	response.Data.HeadSlot = client.Uinteger(highestSlot)
	response.Data.SyncDistance = client.Uinteger(highestSlot - currentSlot)
	return response, nil
}

// ===========
// === NYI ===
// ===========

func (m *BeaconMockManager) Beacon_Attestations(ctx context.Context, blockId string) (client.AttestationsResponse, bool, error) {
	return client.AttestationsResponse{}, false, fmt.Errorf("not implemented")
}

func (m *BeaconMockManager) Beacon_Block(ctx context.Context, blockId string) (client.BeaconBlockResponse, bool, error) {
	return client.BeaconBlockResponse{}, false, fmt.Errorf("not implemented")
}

func (m *BeaconMockManager) Beacon_BlsToExecutionChanges_Post(ctx context.Context, request client.BLSToExecutionChangeRequest) error {
	return fmt.Errorf("not implemented")
}

func (m *BeaconMockManager) Beacon_Committees(ctx context.Context, stateId string, epoch *uint64) (client.CommitteesResponse, error) {
	return client.CommitteesResponse{}, fmt.Errorf("not implemented")
}

func (m *BeaconMockManager) Beacon_FinalityCheckpoints(ctx context.Context, stateId string) (client.FinalityCheckpointsResponse, error) {
	return client.FinalityCheckpointsResponse{}, fmt.Errorf("not implemented")
}

func (m *BeaconMockManager) Beacon_Genesis(ctx context.Context) (client.GenesisResponse, error) {
	return client.GenesisResponse{}, fmt.Errorf("not implemented")
}

func (m *BeaconMockManager) Beacon_Header(ctx context.Context, blockId string) (client.BeaconBlockHeaderResponse, bool, error) {
	return client.BeaconBlockHeaderResponse{}, false, fmt.Errorf("not implemented")
}

func (m *BeaconMockManager) Beacon_VoluntaryExits_Post(ctx context.Context, request client.VoluntaryExitRequest) error {
	return fmt.Errorf("not implemented")
}

func (m *BeaconMockManager) Config_Spec(ctx context.Context) (client.Eth2ConfigResponse, error) {
	return client.Eth2ConfigResponse{}, fmt.Errorf("not implemented")
}

func (m *BeaconMockManager) Validator_DutiesProposer(ctx context.Context, indices []string, epoch uint64) (client.ProposerDutiesResponse, error) {
	return client.ProposerDutiesResponse{}, fmt.Errorf("not implemented")
}

func (m *BeaconMockManager) Validator_DutiesSync_Post(ctx context.Context, indices []string, epoch uint64) (client.SyncDutiesResponse, error) {
	return client.SyncDutiesResponse{}, fmt.Errorf("not implemented")
}
