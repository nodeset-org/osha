package manager

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/nodeset-org/osha/beacon/db"
)

type BeaconManagerModule struct {
	name     string
	database *db.Database

	snapshots map[string]*db.Database
	logger    *slog.Logger
}

func (m *BeaconManagerModule) GetName() string {
	return m.name
}

func (m *BeaconManagerModule) GetRequirements() {
}

func (m *BeaconManagerModule) Close() error {
	if m.database != nil {
		err := m.database.Close()
		if err != nil {
			m.logger.Error("Failed to close database", "error", err)
			return fmt.Errorf("failed to close database: %w", err)
		}
	}

	if m.snapshots != nil {
		for snapshotName := range m.snapshots {
			delete(m.snapshots, snapshotName)
		}
	}

	m.database = nil
	m.snapshots = nil
	return nil
}

func (m *BeaconManagerModule) TakeSnapshot() (string, error) {
	snapshot := m.database.Clone()
	if snapshot == nil {
		return "", fmt.Errorf("failed to clone database for snapshot")
	}
	timestamp := time.Now().Format("20060102_150405")
	snapshotName := fmt.Sprintf("%s_%s", m.name, timestamp)
	m.snapshots[snapshotName] = snapshot

	return snapshotName, nil
}

func (m *BeaconManagerModule) RevertToSnapshot(name string) error {
	snapshot, exists := m.snapshots[name]
	if !exists {
		return fmt.Errorf("snapshot %s does not exist", name)
	}
	m.database = snapshot
	m.logger.Info("Reverted to DB snapshot", "name", name)
	return nil
}

func NewBeaconManagerModule(logger *slog.Logger, config *db.Config) *BeaconManagerModule {
	return &BeaconManagerModule{
		name:      "BeaconManager",
		database:  db.NewDatabase(logger, config.FirstExecutionBlockIndex),
		snapshots: map[string]*db.Database{},
		logger:    logger,
	}
}
