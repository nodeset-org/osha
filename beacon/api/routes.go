package api

const (
	StateID     string = "state_id"
	ValidatorID string = "validator_id"

	// Beacon API routes
	ValidatorsRouteTemplate      string = "v1/beacon/states/%s/validators"
	ValidatorsRoute              string = "v1/beacon/states/{state_id}/validators"
	ValidatorRouteTemplate       string = "v1/beacon/states/%s/validators/%s"
	PendingDepositsRouteTemplate string = "v1/beacon/states/%s/pending_deposits"
	ValidatorRoute               string = "v1/beacon/states/{state_id}/validators/{validator_id}"
	SyncingRoute                 string = "v1/node/syncing"
	PendingDepositsRoute         string = "v1/beacon/states/{state_id}/pending_deposits"
	DepositContractRoute         string = "v1/config/deposit_contract"
	ConfigSpecRoute              string = "v1/config/spec"
	BeaconGenesisRoute           string = "v1/beacon/genesis"
	FinalityCheckpointsRoute     string = "v1/beacon/states/{state_id}/finality_checkpoints"

	// Admin routes
	AddValidatorRoute   string = "add-validator"
	CommitBlockRoute    string = "commit-block"
	SetBalanceRoute     string = "set-balance"
	SetStatusRoute      string = "set-status"
	SetHighestSlotRoute string = "set-highest-slot"
	SlashRoute          string = "slash"
)
