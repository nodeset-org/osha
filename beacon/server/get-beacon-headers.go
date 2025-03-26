package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nodeset-org/osha/beacon/api"
)

// Handle a get beacon genesis request
func (s *BeaconMockServer) getBeaconHeaders(w http.ResponseWriter, r *http.Request) {
	// Get the request vars
	args := s.processApiRequest(w, r, nil)

	slot, exists := args[api.Slot]
	if !exists {
		handleInputError(s.logger, w, fmt.Errorf("missing slot"))
		return
	}

	response, _, err := s.manager.Beacon_Header(context.Background(), slot[0])

	if err != nil {
		handleServerError(s.logger, w, err)
		return
	}
	handleSuccess(s.logger, w, response)
}
