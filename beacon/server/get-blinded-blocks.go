package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nodeset-org/osha/beacon/api"
)

// Handle a get beacon genesis request
func (s *BeaconMockServer) getBlindedBlocks(w http.ResponseWriter, r *http.Request) {
	// Get the request vars
	_ = s.processApiRequest(w, r, nil)

	vars := mux.Vars(r)

	blockID, exists := vars[api.BlockID]
	if !exists {
		handleInputError(s.logger, w, fmt.Errorf("missing block ID"))
		return
	}

	response, _, err := s.manager.Blinded_Block(context.Background(), blockID)

	if err != nil {
		handleServerError(s.logger, w, err)
		return
	}
	handleSuccess(s.logger, w, response)
}
