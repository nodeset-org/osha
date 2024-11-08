package admin

import (
	"net/http"

	"github.com/nodeset-org/osha/vc/api"
	"github.com/nodeset-org/osha/vc/server/common"
)

// Handler for admin/default-fee-recipient
func (s *AdminServer) handleFeeRecipient(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getDefaultFeeRecipient(w, r)
	case http.MethodPost:
		s.setDefaultFeeRecipient(w, r)
	default:
		common.HandleInvalidMethod(w, s.logger)
	}
}

// GET admin/default-fee-recipient
func (s *AdminServer) getDefaultFeeRecipient(w http.ResponseWriter, r *http.Request) {
	db := s.manager.GetDatabase()

	// Write the data
	data := api.GetDefaultFeeRecipientResponse{
		FeeRecipient: db.GetDefaultFeeRecipient(),
	}
	common.HandleSuccess(w, s.logger, data, nil)
}

// POST admin/default-fee-recipient
func (s *AdminServer) setDefaultFeeRecipient(w http.ResponseWriter, r *http.Request) {
	// Get the requesting node
	var body api.SetDefaultFeeRecipientBody
	common.ProcessApiRequest(s, w, r, &body)

	// Input validation
	db := s.manager.GetDatabase()
	db.SetDefaultFeeRecipient(body.FeeRecipient)
	common.HandleSuccess(w, s.logger, struct{}{}, nil)
}
