package admin

import (
	"fmt"
	"net/http"

	"github.com/nodeset-org/osha/vc/server/common"
)

// Revert the server to a previous snapshot
func (s *AdminServer) handleRevert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		common.HandleInvalidMethod(w, s.logger)
		return
	}

	snapshotName := r.URL.Query().Get("name")
	if snapshotName == "" {
		common.HandleInputError(w, s.logger, fmt.Errorf("missing snapshot name"))
		return
	}

	err := s.manager.RevertToSnapshot(snapshotName)
	if err != nil {
		common.HandleServerError(w, s.logger, err)
		return
	}
	common.HandleSuccess(w, s.logger, "", nil)
}
