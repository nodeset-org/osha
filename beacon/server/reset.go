package server

import (
	"net/http"

	"github.com/nodeset-org/osha/beacon/db"
)

func (s *BeaconMockServer) reset(w http.ResponseWriter, r *http.Request) {

	d := db.NewDatabase(s.logger, 0)
	s.manager.SetDatabase(d)

	handleSuccess(s.logger, w, true)
}
