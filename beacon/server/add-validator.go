package server

import (
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/osha/beacon/api"
	"github.com/rocket-pool/node-manager-core/beacon"
)

// Handle an add validator request
func (s *BeaconMockServer) addValidator(w http.ResponseWriter, r *http.Request) {
	// Get the request vars
	args := s.processApiRequest(w, r, nil)
	pubkeyString, exists := args["pubkey"]
	if !exists {
		handleInputError(s.logger, w, fmt.Errorf("missing pubkey"))
		return
	}
	withdrawalCredsString, exists := args["creds"]
	if !exists {
		handleInputError(s.logger, w, fmt.Errorf("missing withdrawal creds"))
		return
	}

	// Input validation
	pubkey, err := beacon.HexToValidatorPubkey(pubkeyString[0])
	if err != nil {
		handleInputError(s.logger, w, fmt.Errorf("invalid pubkey [%s]: %w", pubkeyString[0], err))
		return
	}
	withdrawalCreds := common.HexToHash(withdrawalCredsString[0])

	// Get the validator
	validator, err := s.manager.AddValidator(pubkey, withdrawalCreds)
	if err != nil {
		handleInputError(s.logger, w, err)
		return
	}

	// Respond
	response := api.AddValidatorResponse{
		Index: validator.Index,
	}
	handleSuccess(s.logger, w, response)
}
