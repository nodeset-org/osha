package osha

import (
	"fmt"
	"log/slog"
)

// Interface representing individual module snapshots that compose an entire Snapshot
type IOshaModule interface {
	GetName() string
	GetRequirements()
	Close() error
	TakeSnapshot() (string, error)
	RevertToSnapshot(name string) error
}

// Struct representing an entire snapshot for a given test case
type Snapshot struct {
	name   string
	states map[IOshaModule]any
}

type OshaTestManager struct {
	// logger for logging output messages during tests
	logger *slog.Logger

	// Map of services were captured during a snapshot
	snapshotServiceMap map[string]Snapshot
}

func (tm *OshaTestManager) RegisterModule(module IOshaModule) error {
	// Create a new snapshot for the module and save it in snapshotServiceMap
	snapshot := Snapshot{
		name:   module.GetName(),
		states: make(map[IOshaModule]any),
	}

	// Taking an initial snapshot of the module
	state, err := module.TakeSnapshot()
	if err != nil {
		return fmt.Errorf("failed to take snapshot for module %s: %w", module.GetName(), err)
	}
	snapshot.states[module] = state

	// Store the snapshot in the snapshotServiceMap
	tm.snapshotServiceMap[module.GetName()] = snapshot
	return nil
}

func (tm *OshaTestManager) CreateCustomSnapshot() (string, error) {
	// Create a new snapshot
	snapshot := Snapshot{
		name:   "snapshot_" + fmt.Sprint(len(tm.snapshotServiceMap)+1), // TODO: snapshot format?
		states: make(map[IOshaModule]any),
	}

	// Take a snapshot of each registered module
	for _, existingSnapshot := range tm.snapshotServiceMap {
		for module, state := range existingSnapshot.states {
			snapshot.states[module] = state
		}
	}

	tm.snapshotServiceMap[snapshot.name] = snapshot
	return snapshot.name, nil
}

func (tm *OshaTestManager) RevertToSnapshot(name string) error {
	// Check if the snapshot exists
	snapshot, exists := tm.snapshotServiceMap[name]
	if !exists {
		return fmt.Errorf("snapshot %s does not exist", name)
	}

	// Revert each module to the state in the snapshot (all or nothing)
	for module, state := range snapshot.states {
		// Revert the module's state to the stored snapshot
		err := module.RevertToSnapshot(state.(string))
		if err != nil {
			return fmt.Errorf("failed to revert module %s: %w", module.GetName(), err)
		}
	}

	return nil
}
