package admin

import (
	"net/http"

	"github.com/nodeset-org/osha/vc/api"
	"github.com/nodeset-org/osha/vc/server/common"
)

// Handler for admin/default-graffiti
func (s *AdminServer) handleGraffiti(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getDefaultGraffiti(w, r)
	case http.MethodPost:
		s.setDefaultGraffiti(w, r)
	default:
		common.HandleInvalidMethod(w, s.logger)
	}
}

// GET admin/default-graffiti
func (s *AdminServer) getDefaultGraffiti(w http.ResponseWriter, r *http.Request) {
	db := s.manager.GetDatabase()

	// Write the data
	data := api.GetDefaultGraffitiResponse{
		Graffiti: db.GetDefaultGraffiti(),
	}
	common.HandleSuccess(w, s.logger, data, nil)
}

// POST admin/default-graffiti
func (s *AdminServer) setDefaultGraffiti(w http.ResponseWriter, r *http.Request) {
	// Get the requesting node
	var body api.SetDefaultGraffitiBody
	common.ProcessApiRequest(s, w, r, &body)

	// Input validation
	db := s.manager.GetDatabase()
	db.SetDefaultGraffiti(body.Graffiti)
	common.HandleSuccess(w, s.logger, struct{}{}, nil)
}
