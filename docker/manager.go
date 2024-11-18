package docker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type DockerMockManager struct {
	client.APIClient

	name string

	// Internal fields
	state     *state
	snapshots map[string]*state
	logger    *slog.Logger
}

func (m *DockerMockManager) GetName() string {
	return m.name
}

func (m *DockerMockManager) GetRequirements() {
}

// Creates a new Docker mock manager instance
func NewDockerMockManager(logger *slog.Logger) *DockerMockManager {
	return &DockerMockManager{
		name:      "DockerMockManager",
		state:     newState(),
		snapshots: map[string]*state{},
		logger:    logger,
	}
}

// Creates a new snapshot of the current state of the Docker mock manager
func (m *DockerMockManager) TakeSnapshot() error {
	// Clone the state
	clone, err := m.state.Clone()
	if err != nil {
		return err
	}
	timestamp := time.Now().Format("20060102_150405")
	snapshotName := fmt.Sprintf("%s_%s", m.name, timestamp)

	// Store the snapshot
	m.snapshots[snapshotName] = clone
	m.logger.Info("Took Docker snapshot", "name", snapshotName)
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

func (m *DockerMockManager) Close() error {
	// Remove network
	for networkID := range m.state.networks {
		err := m.NetworkRemove(context.Background(), networkID)
		if err != nil {
			m.logger.Error("Failed to remove network", "error", err)
		}
	}

	// Stop running containers, remove any network references
	for containerID, containerjson := range m.state.containers {
		if containerjson.State.Running {
			err := m.ContainerStop(context.Background(), containerID, container.StopOptions{})
			if err != nil {
				m.logger.Error("Failed to stop container", "error", err)
			}
		}

		if containerjson.NetworkSettings != nil {
			for networkName := range containerjson.NetworkSettings.Networks {
				network, exists := m.state.networks[networkName]
				if exists {
					delete(network.Containers, containerjson.ID)
				}
			}
		}
		err := m.ContainerRemove(context.Background(), containerID, container.RemoveOptions{})
		if err != nil {
			m.logger.Error("Failed to remove container", "error", err)
		}
	}

	// Remove volumes
	for volumeID := range m.state.volumes {
		err := m.VolumeRemove(context.Background(), volumeID, false)
		if err != nil {
			m.logger.Error("Failed to remove volume", "error", err)
		}
	}

	if m.snapshots != nil {
		for snapshotName := range m.snapshots {
			delete(m.snapshots, snapshotName)
		}
	}

	m.state = nil
	m.snapshots = nil

	return nil
}
