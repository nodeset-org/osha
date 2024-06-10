package api

import (
	"github.com/rocket-pool/node-manager-core/beacon/client"
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
