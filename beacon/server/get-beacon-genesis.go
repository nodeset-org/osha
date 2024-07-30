package server

import (
	"context"
	"net/http"
)

// Handle a get beacon genesis request
func (s *BeaconMockServer) getBeaconGenesis(w http.ResponseWriter, r *http.Request) {
	// Get the request vars
	_ = s.processApiRequest(w, r, nil)
	response, err := s.manager.Beacon_Genesis(context.Background())
	if err != nil {
		handleServerError(s.logger, w, err)
		return
	}
	handleSuccess(s.logger, w, response)
}
