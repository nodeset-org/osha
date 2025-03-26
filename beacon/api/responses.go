package api

import (
	"github.com/rocket-pool/node-manager-core/beacon/client"
	"github.com/rocket-pool/node-manager-core/utils"
)

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ValidatorResponse struct {
	Data client.Validator `json:"data"`
}

type AddValidatorResponse struct {
	Index uint64 `json:"index"`
}

type BlindedBlockResponse struct {
	Data struct {
		Message struct {
			Slot          utils.Uinteger `json:"slot"`
			ProposerIndex string         `json:"proposer_index"`
			Body          struct {
				Eth1Data struct {
					DepositRoot  utils.ByteArray `json:"deposit_root"`
					DepositCount utils.Uinteger  `json:"deposit_count"`
					BlockHash    utils.ByteArray `json:"block_hash"`
				} `json:"eth1_data"`
				Attestations           []client.Attestation `json:"attestations"`
				ExecutionPayloadHeader struct {
					FeeRecipient utils.ByteArray `json:"fee_recipient"`
					BlockNumber  utils.Uinteger  `json:"block_number"`
				} `json:"execution_payload"`
			} `json:"body"`
		} `json:"message"`
	} `json:"data"`
}
