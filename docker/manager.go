package docker

import (
	"fmt"
	"log/slog"

	"github.com/docker/docker/client"
)

type DockerMockManager struct {
	client.APIClient

	// Internal fields
	state     *state
	snapshots map[string]*state
	logger    *slog.Logger
}

// Creates a new Docker mock manager instance
func NewDockerMockManager(logger *slog.Logger) *DockerMockManager {
	return &DockerMockManager{
		state:     newState(),
		snapshots: map[string]*state{},
		logger:    logger,
	}
}

// Creates a new snapshot of the current state of the Docker mock manager
func (m *DockerMockManager) TakeSnapshot(name string) error {
	// Clone the state
	clone, err := m.state.Clone()
	if err != nil {
		return err
	}

	// Store the snapshot
	m.snapshots[name] = clone
	m.logger.Info("Took Docker snapshot", "name", name)
	return nil
}

// Revert to a snapshot of the Docker mock state
func (m *DockerMockManager) RevertToSnapshot(name string) error {
	clone, exists := m.snapshots[name]
	if !exists {
		return fmt.Errorf("snapshot with name [%s] does not exist", name)
	}
	m.state = clone
	m.logger.Info("Reverted to Docker snapshot", "name", name)
	return nil
}
