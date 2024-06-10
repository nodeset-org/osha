package server

import (
	"context"
	"net/http"
)

// Handle a get sync status request
func (s *BeaconMockServer) getSyncStatus(w http.ResponseWriter, r *http.Request) {
	// Get the request vars
	_ = s.processApiRequest(w, r, nil)
	response, err := s.manager.Node_Syncing(context.Background())
	if err != nil {
		handleServerError(s.logger, w, err)
		return
	}
	handleSuccess(s.logger, w, response)
}
