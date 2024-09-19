package server

import (
	"context"
	"net/http"
)

// Handle a get deposit contract request
func (s *BeaconMockServer) getDepositContract(w http.ResponseWriter, r *http.Request) {
	// Get the request vars
	_ = s.processApiRequest(w, r, nil)
	response, err := s.manager.Config_DepositContract(context.Background())
	if err != nil {
		handleServerError(s.logger, w, err)
		return
	}
	handleSuccess(s.logger, w, response)
}
