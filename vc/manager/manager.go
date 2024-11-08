package manager

import (
	"fmt"
	"log/slog"

	"github.com/nodeset-org/osha/vc/db"
)

// Mock manager for the VC keymanager API
type VcMockManager struct {
	database *db.KeyManagerDatabase

	// Internal fields
	snapshots map[string]*db.KeyManagerDatabase
	logger    *slog.Logger
}

// Creates a new manager
func NewVcMockManager(logger *slog.Logger, dbOpts db.KeyManagerDatabaseOptions) *VcMockManager {
	return &VcMockManager{
		database:  db.NewKeyManagerDatabase(logger, dbOpts),
		snapshots: map[string]*db.KeyManagerDatabase{},
		logger:    logger,
	}
}

// Get the database the manager is currently using
func (m *VcMockManager) GetDatabase() *db.KeyManagerDatabase {
	return m.database
}

// Set the database for the manager directly if you need to custom provision it
func (m *VcMockManager) SetDatabase(db *db.KeyManagerDatabase) {
	m.database = db
}

// Take a snapshot of the current database state
func (m *VcMockManager) TakeSnapshot(name string) {
	m.snapshots[name] = m.database.Clone()
	m.logger.Info("Took DB snapshot", "name", name)
}

// Revert to a snapshot of the database state
func (m *VcMockManager) RevertToSnapshot(name string) error {
	snapshot, exists := m.snapshots[name]
	if !exists {
		return fmt.Errorf("snapshot with name [%s] does not exist", name)
	}
	m.database = snapshot
	m.logger.Info("Reverted to DB snapshot", "name", name)
	return nil
}
