package api

const (
	PubkeyID string = "pubkey"

	// Keymanager API routes
	FeeRecipientRoute string = "v1/validator/{pubkey}/feerecipient"
	GraffitiRoute     string = "v1/validator/{pubkey}/graffiti"
	KeystoresRoute    string = "v1/keystores"

	// Admin routes
	AdminDefaultGraffitiRoute       string = "default-graffiti"
	AdminDefaultFeeRecipientRoute   string = "default-fee-recipient"
	AdminGenesisValidatorsRootRoute string = "genesis-validators-root"
	AdminJwtSecretRoute             string = "jwt-secret"
	AdminSnapshotRoute              string = "snapshot"
	AdminRevertRoute                string = "revert"
)
