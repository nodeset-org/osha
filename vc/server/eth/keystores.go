package eth

import (
	"encoding/json"
	"net/http"

	"github.com/nodeset-org/osha/vc/server/common"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/node/validator/keymanager"
)

// Handler for eth/v1/keystores
func (s *KeyManagerServer) handleKeystores(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getKeystores(w, r)
	case http.MethodPost:
		s.setKeystores(w, r)
	case http.MethodDelete:
		s.deleteKeystores(w, r)
	default:
		common.HandleInvalidMethod(w, s.logger)
	}
}

// GET eth/v1/keystores
func (s *KeyManagerServer) getKeystores(w http.ResponseWriter, r *http.Request) {
	common.ProcessApiRequest(s, w, r, nil)
	if !common.ProcessAuthHeader(s, w, r) {
		return
	}

	// Write the data
	db := s.manager.GetDatabase()
	data := db.GetAllValidators()
	common.HandleSuccess(w, s.logger, data, nil)
}

// POST eth/v1/keystores
func (s *KeyManagerServer) setKeystores(w http.ResponseWriter, r *http.Request) {
	var body keymanager.ImportKeysBody
	_, _, success := common.ProcessApiRequest(s, w, r, &body)
	if !success {
		return
	}
	if !common.ProcessAuthHeader(s, w, r) {
		return
	}

	// Prep the body
	keystores := make([]*beacon.ValidatorKeystore, len(body.Keystores))
	for i, marshalledKeystore := range body.Keystores {
		var keystore beacon.ValidatorKeystore
		err := json.Unmarshal([]byte(marshalledKeystore), &keystore)
		if err != nil {
			common.HandleInputError(w, s.logger, err)
			return
		}
		keystores[i] = &keystore
	}
	var slashingProtection *beacon.SlashingProtectionData
	if body.SlashingProtection != "" {
		slashingProtection = new(beacon.SlashingProtectionData)
		err := json.Unmarshal([]byte(body.SlashingProtection), slashingProtection)
		if err != nil {
			common.HandleInputError(w, s.logger, err)
			return
		}
	}

	// Write the data
	db := s.manager.GetDatabase()

	data, err := db.AddValidators(keystores, body.Passwords, slashingProtection)
	if err != nil {
		common.HandleServerError(w, s.logger, err)
		return
	}
	common.HandleSuccess(w, s.logger, data, nil)
}

// DELETE eth/v1/keystores
func (s *KeyManagerServer) deleteKeystores(w http.ResponseWriter, r *http.Request) {
	var body keymanager.DeleteKeysBody
	_, _, success := common.ProcessApiRequest(s, w, r, &body)
	if !success {
		return
	}
	if !common.ProcessAuthHeader(s, w, r) {
		return
	}

	// Write the data
	db := s.manager.GetDatabase()
	data, slashingProtection := db.DeleteValidators(body.Pubkeys)
	common.HandleSuccess(w, s.logger, data, slashingProtection)
}
