package admin

import (
	"net/http"

	"github.com/nodeset-org/osha/vc/api"
	"github.com/nodeset-org/osha/vc/server/common"
)

// Handler for admin/genesis-validators-root
func (s *AdminServer) handleGenesisValidatorsRoot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getGenesisValidatorsRoot(w, r)
	case http.MethodPost:
		s.setGenesisValidatorsRoot(w, r)
	default:
		common.HandleInvalidMethod(w, s.logger)
	}
}

// GET admin/genesis-validators-root
func (s *AdminServer) getGenesisValidatorsRoot(w http.ResponseWriter, r *http.Request) {
	db := s.manager.GetDatabase()

	// Write the data
	data := api.GetGenesisValidatorsRootResponse{
		Root: db.GetGenesisValidatorsRoot(),
	}
	common.HandleSuccess(w, s.logger, data, nil)
}

// POST admin/genesis-validators-root
func (s *AdminServer) setGenesisValidatorsRoot(w http.ResponseWriter, r *http.Request) {
	// Get the requesting node
	var body api.SetGenesisValidatorsRootBody
	common.ProcessApiRequest(s, w, r, &body)

	// Input validation
	db := s.manager.GetDatabase()
	db.SetGenesisValidatorsRoot(body.Root)
	common.HandleSuccess(w, s.logger, struct{}{}, nil)
}
