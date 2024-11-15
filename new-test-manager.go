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

	// Taking a snapshot of the module when it's registered
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
		name:   "snapshot_" + fmt.Sprint(len(tm.snapshotServiceMap)+1), // Unique snapshot name
		states: make(map[IOshaModule]any),
	}

	// Take a snapshot of each registered module (from snapshotServiceMap)
	for _, existingSnapshot := range tm.snapshotServiceMap {
		for module, state := range existingSnapshot.states {
			// Store the module state in the new snapshot
			snapshot.states[module] = state
		}
	}

	// Save the snapshot in the map
	tm.snapshotServiceMap[snapshot.name] = snapshot

	// Return the snapshot name
	return snapshot.name, nil
}

func (tm *OshaTestManager) RevertToSnapshot(name string) error {
	// Check if the snapshot exists
	snapshot, exists := tm.snapshotServiceMap[name]
	if !exists {
		return fmt.Errorf("snapshot %s does not exist", name)
	}

	// Revert each module to the state in the snapshot
	for module, state := range snapshot.states {
		// Revert the module's state to the stored snapshot
		err := module.RevertToSnapshot(state.(string)) // Assuming state is of type string
		if err != nil {
			return fmt.Errorf("failed to revert module %s: %w", module.GetName(), err)
		}
	}

	return nil
}
