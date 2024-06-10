package server

import (
	"fmt"
	"net/http"
	"strconv"
)

func (s *BeaconMockServer) setBalance(w http.ResponseWriter, r *http.Request) {
	// Get the request vars
	args := s.processApiRequest(w, r, nil)
	id, exists := args["id"]
	if !exists {
		handleInputError(s.logger, w, fmt.Errorf("missing validator ID"))
		return
	}
	balanceString, exists := args["balance"]
	if !exists {
		handleInputError(s.logger, w, fmt.Errorf("missing balance"))
		return
	}

	// Input validation
	balance, err := strconv.ParseUint(balanceString[0], 10, 64)
	if err != nil {
		handleInputError(s.logger, w, fmt.Errorf("invalid balance [%s]: %w", balanceString[0], err))
		return
	}

	// Get the validator
	validator, err := s.manager.GetValidator(id[0])
	if err != nil {
		handleInputError(s.logger, w, err)
		return
	}

	// Set the balance
	validator.SetBalance(balance)
	handleSuccess(s.logger, w, nil)
}
