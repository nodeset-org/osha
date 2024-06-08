package docker

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	containerimpl "github.com/docker/docker/container"
	"github.com/stretchr/testify/require"
)

func TestContainerCycle(t *testing.T) {
	// Create a new client
	d := NewDockerClientMock()

	// Make a simple service
	service := createTestService()
	err := d.Mock_AddContainer(*service)
	if err != nil {
		t.Fatalf("error adding container: %s", err)
	}
	require.Contains(t, d.containers, service.Name)
	t.Log("Added container")

	// Start the service
	err = d.ContainerStart(context.Background(), service.Name, container.StartOptions{})
	if err != nil {
		t.Fatalf("error starting container: %s", err)
	}
	require.True(t, d.containers[service.Name].State.Running)
	t.Log("Started container")

	// Stop the service
	err = d.ContainerStop(context.Background(), service.Name, container.StopOptions{})
	if err != nil {
		t.Fatalf("error stopping container: %s", err)
	}
	require.False(t, d.containers[service.Name].State.Running)
	t.Log("Stopped container")

	// Restart the service
	err = d.ContainerRestart(context.Background(), service.Name, container.StopOptions{})
	if err != nil {
		t.Fatalf("error restarting container: %s", err)
	}
	require.True(t, d.containers[service.Name].State.Running)
	t.Log("Restarted container")

	// Remove the service
	err = d.ContainerRemove(context.Background(), service.Name, container.RemoveOptions{})
	if err != nil {
		t.Fatalf("error removing container: %s", err)
	}
	require.NotContains(t, d.containers, service.Name)
	t.Log("Removed container")
}

func createTestService() *types.ContainerJSON {
	state := containerimpl.NewState()
	stateImpl := getContainerStateFromState(state)

	return &types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			ID:      createRandomID(32),
			Created: time.Now().Format(time.RFC3339Nano),
			Path:    "/usr/bin/dummy",
			Name:    "test",
			Image:   "mock/test:v0.0.1",
			Args:    []string{"arg1", "arg2"},
			HostConfig: &container.HostConfig{
				RestartPolicy: container.RestartPolicy{
					Name:              container.RestartPolicyUnlessStopped,
					MaximumRetryCount: 0,
				},
			},
			State: &stateImpl,
		},
	}
}
