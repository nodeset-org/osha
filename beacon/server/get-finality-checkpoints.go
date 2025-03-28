package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nodeset-org/osha/beacon/api"
)

// Handle a get finality checkpoints request
func (s *BeaconMockServer) getFinalityCheckpoints(w http.ResponseWriter, r *http.Request) {
	// Get the request vars
	vars := mux.Vars(r)
	state, exists := vars[api.StateID]
	if !exists {
		handleInputError(s.logger, w, fmt.Errorf("missing state ID"))
		return
	}
	/*
		if state != "head" {
			handleInputError(s.logger, w, fmt.Errorf("unsupported state ID [%s], only 'head' is supported", state))
			return
		}
	*/

	// Get the response
	response, err := s.manager.Beacon_FinalityCheckpoints(context.Background(), state)
	if err != nil {
		handleInputError(s.logger, w, err)
		return
	}
	handleSuccess(s.logger, w, response)
}
