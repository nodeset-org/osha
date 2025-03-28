package server

import (
	"fmt"
	"net/http"
	"strconv"
)

func (s *BeaconMockServer) setActivationEpoch(w http.ResponseWriter, r *http.Request) {
	// Get the request vars
	args := s.processApiRequest(w, r, nil)
	id, exists := args["id"]
	if !exists {
		handleInputError(s.logger, w, fmt.Errorf("missing validator ID"))
		return
	}
	epochString, exists := args["epoch"]
	if !exists {
		handleInputError(s.logger, w, fmt.Errorf("missing epoch"))
		return
	}

	// Input validation

	epoch, err := strconv.ParseUint(epochString[0], 10, 64)
	if err != nil {
		handleInputError(s.logger, w, fmt.Errorf("invalid epoch [%s]", epochString[0]))
		return

	}

	// Get the validator
	validator, err := s.manager.GetValidator(id[0])
	if err != nil {
		handleInputError(s.logger, w, err)
		return
	}

	// Set the status
	validator.SetActivationEpoch(epoch)
	handleSuccess(s.logger, w, nil)
}
