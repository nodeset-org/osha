package eth

import (
	"fmt"
	"net/http"

	"github.com/nodeset-org/osha/vc/server/common"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/node/validator/keymanager"
)

// Handler for eth/v1/validators/{pubkey}/graffiti
func (s *KeyManagerServer) handleGraffiti(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getGraffiti(w, r)
	case http.MethodPost:
		s.setGraffiti(w, r)
	default:
		common.HandleInvalidMethod(w, s.logger)
	}
}

// GET eth/v1/validators/{pubkey}/graffiti
func (s *KeyManagerServer) getGraffiti(w http.ResponseWriter, r *http.Request) {
	_, pathArgs, success := common.ProcessApiRequest(s, w, r, nil)
	if !success {
		return
	}
	if !common.ProcessAuthHeader(s, w, r) {
		return
	}

	// Input validation
	pubkeyString := pathArgs["pubkey"]
	pubkey, err := beacon.HexToValidatorPubkey(pubkeyString)
	if err != nil {
		common.HandleInputError(w, s.logger, err)
		return
	}

	// Write the data
	db := s.manager.GetDatabase()
	data := db.GetGraffiti(pubkey)
	common.HandleSuccess(w, s.logger, data, nil)
}

// POST eth/v1/validators/{pubkey}/graffiti
func (s *KeyManagerServer) setGraffiti(w http.ResponseWriter, r *http.Request) {
	var body keymanager.SetGraffitiBody
	_, pathArgs, success := common.ProcessApiRequest(s, w, r, &body)
	if !success {
		return
	}
	if !common.ProcessAuthHeader(s, w, r) {
		return
	}

	// Input validation
	pubkeyString := pathArgs["pubkey"]
	pubkey, err := beacon.HexToValidatorPubkey(pubkeyString)
	if err != nil {
		common.HandleInputError(w, s.logger, err)
		return
	}

	// Write the data
	db := s.manager.GetDatabase()
	if db.SetGraffiti(pubkey, body.Graffiti) {
		common.HandleAccepted(w, s.logger)
	} else {
		common.HandleInputError(w, s.logger, fmt.Errorf("validator not found"))
	}
}
