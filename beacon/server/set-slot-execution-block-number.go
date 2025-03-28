package server

import (
	"fmt"
	"net/http"
	"strconv"
)

// Handle an add validator request
func (s *BeaconMockServer) setSlotExecutionBlockNumber(w http.ResponseWriter, r *http.Request) {
	// Get the request vars
	args := s.processApiRequest(w, r, nil)
	slotString, exists := args["slot"]
	if !exists {
		handleInputError(s.logger, w, fmt.Errorf("missing slot"))
		return
	}

	slot, err := strconv.ParseUint(slotString[0], 10, 64)

	if err != nil {
		handleInputError(s.logger, w, fmt.Errorf("invalid slot [%s]: %w", slotString[0], err))
		return
	}

	blockNumberString, exists := args["block_number"]

	if !exists {
		handleInputError(s.logger, w, fmt.Errorf("missing block number"))
		return
	}

	blockNumber, err := strconv.ParseUint(blockNumberString[0], 10, 64)

	if err != nil {
		handleInputError(s.logger, w, fmt.Errorf("invalid block number [%s]: %w", blockNumberString[0], err))
		return
	}

	// Get the validator
	s.manager.SetSlotExecutionBlockNumber(slot, blockNumber)

	handleSuccess(s.logger, w, nil)
}
