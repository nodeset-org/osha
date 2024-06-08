package server

import (
	"fmt"
	"net/http"
	"strconv"
)

func (s *BeaconMockServer) slash(w http.ResponseWriter, r *http.Request) {
	// Get the request vars
	args := s.processApiRequest(w, r, nil)
	id, exists := args["id"]
	if !exists {
		handleInputError(s.logger, w, fmt.Errorf("missing validator ID"))
		return
	}
	penaltyString, exists := args["penalty"]
	if !exists {
		handleInputError(s.logger, w, fmt.Errorf("missing penalty"))
		return
	}

	// Input validation
	penalty, err := strconv.ParseUint(penaltyString[0], 10, 64)
	if err != nil {
		handleInputError(s.logger, w, fmt.Errorf("invalid penalty [%s]: %w", penaltyString[0], err))
		return
	}

	// Get the validator
	validator, err := s.manager.GetValidator(id[0])
	if err != nil {
		handleInputError(s.logger, w, err)
		return
	}

	// Slash the validator
	err = validator.Slash(penalty)
	if err != nil {
		handleInputError(s.logger, w, err)
		return
	}
	handleSuccess(s.logger, w, nil)
}
