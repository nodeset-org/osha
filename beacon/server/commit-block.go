package server

import (
	"fmt"
	"net/http"
	"strconv"
)

func (s *BeaconMockServer) commitBlock(w http.ResponseWriter, r *http.Request) {
	// Get the request vars
	args := s.processApiRequest(w, r, nil)
	validatedString, exists := args["validated"]
	if !exists {
		handleInputError(s.logger, w, fmt.Errorf("missing validated arg"))
		return
	}

	// Input validation
	validated, err := strconv.ParseBool(validatedString[0])
	if err != nil {
		handleInputError(s.logger, w, fmt.Errorf("error parsing validated arg [%s]: %w", validatedString[0], err))
		return
	}

	// Set the slot
	s.manager.CommitBlock(validated)
	handleSuccess(s.logger, w, nil)
}
