package api

import "github.com/ethereum/go-ethereum/common"

// ================
// === Requests ===
// ================

type SetDefaultFeeRecipientBody struct {
	FeeRecipient common.Address `json:"feeRecipient"`
}

type SetDefaultGraffitiBody struct {
	Graffiti string `json:"graffiti"`
}

type SetGenesisValidatorsRootBody struct {
	Root common.Hash `json:"root"`
}

type SetJwtSecretBody struct {
	Secret string `json:"secret"`
}

// =================
// === Responses ===
// =================

type GetDefaultFeeRecipientResponse struct {
	FeeRecipient common.Address `json:"feeRecipient"`
}

type GetDefaultGraffitiResponse struct {
	Graffiti string `json:"graffiti"`
}

type GetGenesisValidatorsRootResponse struct {
	Root common.Hash `json:"root"`
}

type GetJwtSecretResponse struct {
	Secret string `json:"secret"`
}
