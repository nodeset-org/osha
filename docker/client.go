package docker

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	containerimpl "github.com/docker/docker/container"
)

type DockerClientMock struct {
	client.APIClient
	containers map[string]*types.ContainerJSON
	volumes    map[string]*volume.Volume
	networks   map[string]*types.NetworkResource

	// Internal fields
	availableSubnets []int
	usedSubnets      map[string]int

	// Used to create sequential IP / MAC addresses for services
	networkIndices map[string]byte
}

func NewDockerClientMock() *DockerClientMock {
	// Docker defaults to 31 available subnets for bridges: 172.17.0.0/16 through 172.31.0.0/16
	availableSubnets := []int{}
	for i := 17; i < 32; i++ {
		availableSubnets = append(availableSubnets, i)
	}

	return &DockerClientMock{
		containers: map[string]*types.ContainerJSON{},
		volumes:    map[string]*volume.Volume{},
		networks:   map[string]*types.NetworkResource{},

		availableSubnets: availableSubnets,
		usedSubnets:      map[string]int{},
		networkIndices:   map[string]byte{},
	}
}

// =====================
// === API Functions ===
// =====================

// Returns information about a container.
func (d *DockerClientMock) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	// Find the container
	container, exists := d.containers[containerID]
	if !exists {
		//lint:ignore ST1005 This is what the actual Docker daemon returns
		return types.ContainerJSON{}, fmt.Errorf("No such container: %s", containerID)
	}

	return *container, nil
}

// Lists all running Docker containers (unless options.All is true, in which case lists all containers).
// The other options are not implemented.
func (d *DockerClientMock) ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
	containers := []types.Container{}
	for _, containerJson := range d.containers {
		if options.All || containerJson.State.Running {
			container, err := convertContainerJsonToContainer(containerJson)
			if err != nil {
				return nil, fmt.Errorf("error getting details for container [%s]: %w", containerJson.Name, err)
			}
			containers = append(containers, container)
		}
	}
	return containers, nil
}

// Removes a container.
// The opts are not implemented.
func (d *DockerClientMock) ContainerRemove(ctx context.Context, containerID string, opts container.RemoveOptions) error {
	// Find the container
	container, exists := d.containers[containerID]
	if !exists {
		//lint:ignore ST1005 This is what the actual Docker daemon returns
		return fmt.Errorf("No such container: %s", containerID)
	}
	delete(d.containers, containerID)

	// Remove the container from any networks it was connected to
	for netName := range container.NetworkSettings.Networks {
		netResource, exists := d.networks[netName]
		if !exists {
			continue
		}
		delete(netResource.Containers, container.ID)
	}
	return nil
}

// Restarts a container.
// The opts are not implemented.
func (d *DockerClientMock) ContainerRestart(ctx context.Context, containerID string, opts container.StopOptions) error {
	return d.ContainerStart(ctx, containerID, container.StartOptions{})
}

// Starts a container.
// The opts are not implemented.
func (d *DockerClientMock) ContainerStart(ctx context.Context, containerID string, opts container.StartOptions) error {
	// Find the container
	container, exists := d.containers[containerID]
	if !exists {
		//lint:ignore ST1005 This is what the actual Docker daemon returns
		return fmt.Errorf("No such container: %s", containerID)
	}

	// Get the state
	state, err := getStateFromContainerState(container.State)
	if err != nil {
		return fmt.Errorf("error getting details for container [%s]: %w", container.Name, err)
	}

	// Stop the container
	state.SetRunning(nil, nil, true)

	// Update the state
	*container.State = getContainerStateFromState(state)
	return nil
}

// Stops a container.
// The opts are not implemented.
func (d *DockerClientMock) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	// Find the container
	container, exists := d.containers[containerID]
	if !exists {
		//lint:ignore ST1005 This is what the actual Docker daemon returns
		return fmt.Errorf("No such container: %s", containerID)
	}

	// Get the state
	state, err := getStateFromContainerState(container.State)
	if err != nil {
		return fmt.Errorf("error getting details for container [%s]: %w", container.Name, err)
	}

	// Stop the container
	state.SetStopped(&containerimpl.ExitStatus{
		ExitCode: 0,
		ExitedAt: time.Now(),
	})

	// Update the state
	*container.State = getContainerStateFromState(state)
	return nil
}

// Gets the current disk usage of each Docker resource.
// These must be explicitly set before calling this with Mock_SetContainerDiskUsage() or Mock_SetVolumeDiskUsage().
// options are not implemented; always returns container and volume info.
func (d *DockerClientMock) DiskUsage(ctx context.Context, options types.DiskUsageOptions) (types.DiskUsage, error) {
	diskUsage := types.DiskUsage{
		Containers: []*types.Container{},
		Volumes:    []*volume.Volume{},
	}
	for _, container := range d.containers {
		containerJson, err := convertContainerJsonToContainer(container)
		if err != nil {
			return types.DiskUsage{}, fmt.Errorf("error getting details for [%s]: %w", container.Name, err)
		}
		diskUsage.Containers = append(diskUsage.Containers, &containerJson)
	}
	for _, volume := range d.volumes {
		diskUsage.Volumes = append(diskUsage.Volumes, volume)
	}
	return diskUsage, nil
}

// Deletes a network if not being used by existing containers.
func (d *DockerClientMock) NetworkRemove(ctx context.Context, networkID string) error {
	// Find the network
	network, exists := d.networks[networkID]
	if !exists {
		//lint:ignore ST1005 This is what the actual Docker daemon returns
		return fmt.Errorf("No such network: %s", networkID)
	}

	// Make sure nothing's using it
	if len(network.Containers) > 0 {
		return fmt.Errorf("network %s is in use still", networkID)
	}

	subnet, exists := d.usedSubnets[networkID]
	if exists {
		d.availableSubnets = append(d.availableSubnets, subnet)
	}
	delete(d.networks, networkID)
	delete(d.usedSubnets, networkID)
	return nil
}

// Deletes a volume.
// Always works, force is not used.
func (d *DockerClientMock) VolumeRemove(ctx context.Context, volumeID string, force bool) error {
	// Find the volume
	_, exists := d.volumes[volumeID]
	if !exists {
		//lint:ignore ST1005 This is what the actual Docker daemon returns
		return fmt.Errorf("No such volume: %s", volumeID)
	}
	delete(d.volumes, volumeID)
	return nil
}

// ============================
// === Mock Admin Functions ===
// ============================

// Adds a container to the mock registry. Note that this does not add any new volumes the container references;
// add those manually with Mock_AddVolume().
func (d *DockerClientMock) Mock_AddContainer(info types.ContainerJSON) error {
	_, exists := d.containers[info.Name]
	if exists {
		return fmt.Errorf("container %s already exists", info.Name)
	}
	d.containers[info.Name] = &info
	return nil
}

// Adds a volume to the mock registry.
func (d *DockerClientMock) Mock_AddVolume(volume volume.Volume) error {
	_, exists := d.volumes[volume.Name]
	if exists {
		return fmt.Errorf("volume %s already exists", volume.Name)
	}
	d.volumes[volume.Name] = &volume
	return nil
}

// Sets a container's disk usage
func (d *DockerClientMock) Mock_SetContainerDiskUsage(containerID string, sizeRootFs int64, sizeRw int64) error {
	// Find the container
	container, exists := d.containers[containerID]
	if !exists {
		return fmt.Errorf("no such container: %s", containerID)
	}
	*container.SizeRootFs = sizeRootFs
	*container.SizeRw = sizeRw
	return nil
}

// Sets a container's disk usage
func (d *DockerClientMock) Mock_SetVolumeDiskUsage(volumeID string, size int64) error {
	// Find the volume
	volume, exists := d.volumes[volumeID]
	if !exists {
		return fmt.Errorf("no such volume: %s", volumeID)
	}
	volume.UsageData.Size = size
	return nil
}

// ==========================
// === Internal Functions ===
// ==========================

// Converts a ContainerJSON to a Container.
func convertContainerJsonToContainer(containerJson *types.ContainerJSON) (types.Container, error) {
	// Parse the creation time
	created, err := time.Parse(time.RFC3339, containerJson.Created)
	if err != nil {
		return types.Container{}, fmt.Errorf("error parsing created time [%s]: %w", containerJson.Created, err)
	}

	// Create the port map string
	ports := []types.Port{}
	for port, portSettings := range containerJson.NetworkSettings.Ports {
		for _, setting := range portSettings {
			publicPort, err := strconv.ParseInt(setting.HostPort, 0, 16)
			if err != nil {
				return types.Container{}, fmt.Errorf("error parsing host port [%s]: %w", setting.HostPort, err)
			}
			ports = append(ports, types.Port{
				IP:          setting.HostIP,
				PrivatePort: uint16(port.Int()),
				PublicPort:  uint16(publicPort),
				Type:        port.Proto(),
			})
		}
	}

	// Get the status string
	state, err := getStateFromContainerState(containerJson.State)
	if err != nil {
		return types.Container{}, fmt.Errorf("error getting status: %w", err)
	}

	return types.Container{
		ID:         containerJson.ID,
		Names:      []string{containerJson.Name},
		Image:      containerJson.Config.Image,
		ImageID:    containerJson.Image,
		Command:    fmt.Sprintf("%s %s", containerJson.Path, strings.Join(containerJson.Args, " ")),
		Created:    created.Unix(),
		Ports:      ports,
		SizeRw:     *containerJson.SizeRw,
		SizeRootFs: *containerJson.SizeRootFs,
		Labels:     containerJson.Config.Labels,
		State:      state.StateString(),
		Status:     state.String(),
		HostConfig: struct {
			NetworkMode string "json:\",omitempty\""
		}{
			NetworkMode: string(containerJson.HostConfig.NetworkMode),
		},
		Mounts: containerJson.Mounts,
		NetworkSettings: &types.SummaryNetworkSettings{
			Networks: containerJson.NetworkSettings.Networks,
		},
	}, nil
}

// Creates a State implementation from a ContainerJSON state.
func getStateFromContainerState(containerState *types.ContainerState) (*containerimpl.State, error) {
	// Parse the start time
	startedAt, err := time.Parse(time.RFC3339, containerState.StartedAt)
	if err != nil {
		return nil, fmt.Errorf("error parsing start time [%s]: %w", containerState.StartedAt, err)
	}

	// Parse the stop time
	finishedAt, err := time.Parse(time.RFC3339, containerState.FinishedAt)
	if err != nil {
		return nil, fmt.Errorf("error parsing stop time [%s]: %w", containerState.FinishedAt, err)
	}

	state := &containerimpl.State{
		Running:       containerState.Running,
		Paused:        containerState.Paused,
		Restarting:    containerState.Restarting,
		OOMKilled:     containerState.OOMKilled,
		Dead:          containerState.Dead,
		Pid:           containerState.Pid,
		ExitCodeValue: containerState.ExitCode,
		ErrorMsg:      containerState.Error,
		StartedAt:     startedAt,
		FinishedAt:    finishedAt,
	}
	if containerState.Health != nil {
		state.Health = &containerimpl.Health{
			Health: *containerState.Health,
		}
	}
	return state, nil
}

// Creates a ContainerJSON state from a State implementation.
func getContainerStateFromState(state *containerimpl.State) types.ContainerState {
	containerState := types.ContainerState{
		Status:     state.String(),
		Running:    state.Running,
		Paused:     state.Paused,
		Restarting: state.Restarting,
		OOMKilled:  state.OOMKilled,
		Dead:       state.Dead,
		Pid:        state.Pid,
		ExitCode:   state.ExitCodeValue,
		Error:      state.ErrorMsg,
		StartedAt:  state.StartedAt.Format(time.RFC3339Nano),
		FinishedAt: state.FinishedAt.Format(time.RFC3339Nano),
	}
	if state.Health != nil {
		containerState.Health = &state.Health.Health
	}
	return containerState
}
