package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/nodeset-org/osha/beacon/api"
)

// Handle a get pending deposits request
func (s *BeaconMockServer) getPendingDeposits(w http.ResponseWriter, r *http.Request) {
	// Get the request vars
	args := s.processApiRequest(w, r, nil)
	if args == nil {
		return
	}
	vars := mux.Vars(r)
	state, exists := vars[api.StateID]
	if !exists {
		handleInputError(s.logger, w, fmt.Errorf("missing state ID"))
		return
	}

	// Get the response
	response, err := s.manager.Beacon_PendingDeposits(context.Background(), state)
	if err != nil {
		handleInputError(s.logger, w, err)
		return
	}
	handleSuccess(s.logger, w, response)
}
