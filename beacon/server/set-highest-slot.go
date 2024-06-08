package server

import (
	"fmt"
	"net/http"
	"strconv"
)

func (s *BeaconMockServer) setHighestSlot(w http.ResponseWriter, r *http.Request) {
	// Get the request vars
	args := s.processApiRequest(w, r, nil)
	slotString, exists := args["slot"]
	if !exists {
		handleInputError(s.logger, w, fmt.Errorf("missing slot"))
		return
	}

	// Input validation
	slot, err := strconv.ParseUint(slotString[0], 10, 64)
	if err != nil {
		handleInputError(s.logger, w, fmt.Errorf("invalid slot [%s]: %w", slotString[0], err))
		return
	}

	// Set the slot
	s.manager.SetHighestSlot(slot)
	handleSuccess(s.logger, w, nil)
}
