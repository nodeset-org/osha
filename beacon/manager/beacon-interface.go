package manager

import (
	"context"

	"github.com/nodeset-org/osha/beacon/api"
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

func (m *BeaconMockManager) Beacon_Header(ctx context.Context, slot_index string) (client.BeaconBlockHeaderResponse, bool, error) {

	response := client.BeaconBlockHeaderResponse{}

	currentSlot := m.database.GetCurrentSlot()

	slot := m.GetSlot(slot_index)

	// Get the block header
	response.Finalized = slot.Index <= currentSlot
	response.Data.Canonical = true
	response.Data.Header.Message.Slot = utils.Uinteger(slot.Index)
	response.Data.Header.Message.ProposerIndex = "0"
	response.Data.Root = slot.BlockRoot.Hex()

	return response, true, nil
}

func (m *BeaconMockManager) Blinded_Block(ctx context.Context, block_id string) (api.BlindedBlockResponse, bool, error) {

	slot := m.GetSlot(block_id)

	if slot == nil {
		return api.BlindedBlockResponse{}, false, nil
	}

	response := api.BlindedBlockResponse{}

	response.Data.Message.ProposerIndex = "0"
	response.Data.Message.Slot = utils.Uinteger(slot.Index)
	response.Data.Message.Body.Eth1Data.DepositRoot = []byte{0x00}
	response.Data.Message.Body.Eth1Data.DepositCount = utils.Uinteger(0)
	response.Data.Message.Body.Eth1Data.BlockHash = []byte{0x00}
	response.Data.Message.Body.Attestations = []client.Attestation{}
	response.Data.Message.Body.ExecutionPayloadHeader.BlockNumber = utils.Uinteger(slot.ExecutionBlockNumber)
	response.Data.Message.Body.ExecutionPayloadHeader.FeeRecipient = []byte{0x00}

	return response, true, nil
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
