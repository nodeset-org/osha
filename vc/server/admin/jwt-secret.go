package admin

import (
	"net/http"

	"github.com/nodeset-org/osha/vc/api"
	"github.com/nodeset-org/osha/vc/server/common"
)

// Handler for admin/jwt-secret
func (s *AdminServer) handleJwtSecret(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getJwtSecret(w, r)
	case http.MethodPost:
		s.setJwtSecret(w, r)
	default:
		common.HandleInvalidMethod(w, s.logger)
	}
}

// GET admin/jwt-secret
func (s *AdminServer) getJwtSecret(w http.ResponseWriter, r *http.Request) {
	db := s.manager.GetDatabase()

	// Write the data
	data := api.GetJwtSecretResponse{
		Secret: db.GetJwtSecret(),
	}
	common.HandleSuccess(w, s.logger, data, nil)
}

// POST admin/jwt-secret
func (s *AdminServer) setJwtSecret(w http.ResponseWriter, r *http.Request) {
	// Get the requesting node
	var body api.SetJwtSecretBody
	common.ProcessApiRequest(s, w, r, &body)

	// Input validation
	db := s.manager.GetDatabase()
	db.SetJwtSecret(body.Secret)
	common.HandleSuccess(w, s.logger, struct{}{}, nil)
}
