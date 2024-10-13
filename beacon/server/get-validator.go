package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nodeset-org/osha/beacon/api"
)

// Handle a get validator request
func (s *BeaconMockServer) getValidator(w http.ResponseWriter, r *http.Request) {
	// Get the request vars
	_ = s.processApiRequest(w, r, nil)
	vars := mux.Vars(r)
	/*
		state, exists := vars[api.StateID]
		if !exists {
			handleInputError(s.logger, w, fmt.Errorf("missing state ID"))
			return
		}
		if state != "head" {
			handleInputError(s.logger, w, fmt.Errorf("unsupported state ID [%s], only 'head' is supported", state))
			return
		}
	*/

	id, exists := vars[api.ValidatorID]
	if !exists {
		handleInputError(s.logger, w, fmt.Errorf("missing validator ID"))
		return
	}

	// Get the validator
	validator, err := s.manager.GetValidator(id)
	if err != nil {
		handleInputError(s.logger, w, err)
		return
	}

	// Write the response
	response := api.ValidatorResponse{
		Data: validator.GetValidatorMeta(),
	}
	handleSuccess(s.logger, w, response)
}
