package api

const (
	// Request Parameters
	StateID     string = "state_id"
	ValidatorID string = "validator_id"
	Slot        string = "slot"
	BlockID     string = "block_id"

	// Beacon API routes
	ValidatorsRouteTemplate    string = "v1/beacon/states/%s/validators"
	ValidatorsRoute            string = "v1/beacon/states/{state_id}/validators"
	ValidatorRouteTemplate     string = "v1/beacon/states/%s/validators/%s"
	ValidatorRoute             string = "v1/beacon/states/{state_id}/validators/{validator_id}"
	SyncingRoute               string = "v1/node/syncing"
	DepositContractRoute       string = "v1/config/deposit_contract"
	ConfigSpecRoute            string = "v1/config/spec"
	BeaconGenesisRoute         string = "v1/beacon/genesis"
	BeaconHeadersRoute         string = "v1/beacon/headers"
	FinalityCheckpointsRoute   string = "v1/beacon/states/{state_id}/finality_checkpoints"
	BlindedBlocksRouteTemplate string = "v1/beacon/blinded_blocks/%s"
	BlindedBlocksRoute         string = "v1/beacon/blinded_blocks/{block_id}"

	// Admin routes
	AddValidatorRoute                string = "add-validator"
	SetSlotBlockRootRoute            string = "set-slot-block-root"
	SetSlotExecutionBlockNumberRoute string = "set-slot-execution-block-number"
	CommitBlockRoute                 string = "commit-block"
	SetBalanceRoute                  string = "set-balance"
	SetStatusRoute                   string = "set-status"
	SetActivationEpochRoute          string = "set-activation-epoch"
	SetHighestSlotRoute              string = "set-highest-slot"
	SlashRoute                       string = "slash"
	ResetRoute                       string = "reset"
)
