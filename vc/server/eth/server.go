package eth

import (
	"log/slog"

	"github.com/gorilla/mux"
	"github.com/nodeset-org/osha/vc/api"
	"github.com/nodeset-org/osha/vc/manager"
)

// Keymanager routes for the server mock
type KeyManagerServer struct {
	logger  *slog.Logger
	manager *manager.VcMockManager
}

// Creates a new key manager server mock
func NewKeyManagerServer(logger *slog.Logger, manager *manager.VcMockManager) *KeyManagerServer {
	return &KeyManagerServer{
		logger:  logger,
		manager: manager,
	}
}

// Gets the logger
func (s *KeyManagerServer) GetLogger() *slog.Logger {
	return s.logger
}

// Gets the manager
func (s *KeyManagerServer) GetManager() *manager.VcMockManager {
	return s.manager
}

// Registers the routes for the server
func (s *KeyManagerServer) RegisterRoutes(adminRouter *mux.Router) {
	adminRouter.HandleFunc("/"+api.FeeRecipientRoute, s.handleFeeRecipient)
	adminRouter.HandleFunc("/"+api.GraffitiRoute, s.handleGraffiti)
	adminRouter.HandleFunc("/"+api.KeystoresRoute, s.handleKeystores)
}
