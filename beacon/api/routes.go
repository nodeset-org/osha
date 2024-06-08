package api

const (
	StateID     string = "state_id"
	ValidatorID string = "validator_id"

	// Beacon API routes
	ValidatorsRouteTemplate string = "v1/beacon/states/%s/validators"
	ValidatorsRoute         string = "v1/beacon/states/{state_id}/validators"
	ValidatorRouteTemplate  string = "v1/beacon/states/%s/validators/%s"
	ValidatorRoute          string = "v1/beacon/states/{state_id}/validators/{validator_id}"
	SyncingRoute            string = "v1/node/syncing"

	// Admin routes
	AddValidatorRoute   string = "add-validator"
	CommitBlockRoute    string = "commit-block"
	SetBalanceRoute     string = "set-balance"
	SetStatusRoute      string = "set-status"
	SetHighestSlotRoute string = "set-highest-slot"
	SlashRoute          string = "slash"
)
