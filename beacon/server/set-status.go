package server

import (
	"fmt"
	"net/http"

	"github.com/rocket-pool/node-manager-core/beacon"
)

func (s *BeaconMockServer) setStatus(w http.ResponseWriter, r *http.Request) {
	// Get the request vars
	args := s.processApiRequest(w, r, nil)
	id, exists := args["id"]
	if !exists {
		handleInputError(s.logger, w, fmt.Errorf("missing validator ID"))
		return
	}
	statusString, exists := args["status"]
	if !exists {
		handleInputError(s.logger, w, fmt.Errorf("missing status"))
		return
	}

	// Input validation
	status := beacon.ValidatorState(statusString[0])
	if status == "" {
		handleInputError(s.logger, w, fmt.Errorf("invalid status [%s]", statusString[0]))
		return

	}

	// Get the validator
	validator, err := s.manager.GetValidator(id[0])
	if err != nil {
		handleInputError(s.logger, w, err)
		return
	}

	// Set the status
	validator.SetStatus(status)
	handleSuccess(s.logger, w, nil)
}
