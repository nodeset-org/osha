package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
)

// Handle an add validator request
func (s *BeaconMockServer) setSlotBlockRoot(w http.ResponseWriter, r *http.Request) {
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

	rootString, exists := args["root"]

	if !exists {
		handleInputError(s.logger, w, fmt.Errorf("missing root hash"))
		return
	}

	root := common.HexToHash(rootString[0])

	// Get the validator
	s.manager.SetSlotBlockRoot(slot, root)

	handleSuccess(s.logger, w, nil)
}
