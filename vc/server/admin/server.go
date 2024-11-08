package admin

import (
	"log/slog"

	"github.com/gorilla/mux"
	"github.com/nodeset-org/osha/vc/api"
	"github.com/nodeset-org/osha/vc/manager"
)

// Admin routes for the server mock
type AdminServer struct {
	logger  *slog.Logger
	manager *manager.VcMockManager
}

// Creates a new API v0 server mock
func NewAdminServer(logger *slog.Logger, manager *manager.VcMockManager) *AdminServer {
	return &AdminServer{
		logger:  logger,
		manager: manager,
	}
}

// Gets the logger
func (s *AdminServer) GetLogger() *slog.Logger {
	return s.logger
}

// Gets the manager
func (s *AdminServer) GetManager() *manager.VcMockManager {
	return s.manager
}

// Registers the routes for the server
func (s *AdminServer) RegisterRoutes(adminRouter *mux.Router) {
	adminRouter.HandleFunc("/"+api.AdminDefaultFeeRecipientRoute, s.handleFeeRecipient)
	adminRouter.HandleFunc("/"+api.AdminDefaultGraffitiRoute, s.handleGraffiti)
	adminRouter.HandleFunc("/"+api.AdminGenesisValidatorsRootRoute, s.handleGenesisValidatorsRoot)
	adminRouter.HandleFunc("/"+api.AdminJwtSecretRoute, s.handleJwtSecret)
	adminRouter.HandleFunc("/"+api.AdminSnapshotRoute, s.handleSnapshot)
	adminRouter.HandleFunc("/"+api.AdminRevertRoute, s.handleRevert)
}
