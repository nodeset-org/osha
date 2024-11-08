package common

import (
	"log/slog"

	"github.com/nodeset-org/osha/vc/manager"
)

// Interface for the server mock implementation
type IServerImpl interface {
	// Get the logger
	GetLogger() *slog.Logger

	// Get the manager
	GetManager() *manager.VcMockManager
}
