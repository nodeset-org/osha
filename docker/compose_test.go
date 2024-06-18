package docker

import (
	"fmt"
	"testing"

	"github.com/docker/docker/api/types"
	dcontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	dnetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
)

const (
	projectName       string = "dummyproj"
	goodComposeFile   string = "./test_files/service-good.yml"
	goodContainerName string = "test"
)

func TestComposeUp(t *testing.T) {
	// Create a new client
	netName := fmt.Sprintf("%s_net", projectName)
	volumeName := fmt.Sprintf("%s_named-vol", projectName)
	containerName := fmt.Sprintf("%s_%s", projectName, goodContainerName)
	d := NewDockerMockManager(nil)

	// Run an up command
	err := d.Up(projectName, []string{goodComposeFile})
	if err != nil {
		t.Fatalf("error running compose up: %s", err)
	}
	t.Log("Compose up complete")

	// Verify the network is created
	require.Len(t, d.state.networks, 2)
	require.Contains(t, d.state.networks, netName)

	network := d.state.networks[netName]
	require.Equal(t, netName, network.Name)
	require.Equal(t, "172.18.0.0/16", network.IPAM.Config[0].Subnet)
	require.Equal(t, "172.18.0.1", network.IPAM.Config[0].Gateway)
	require.Len(t, network.Containers, 1)
	t.Log("Network created properly")

	// Verify the volume is created
	require.Len(t, d.state.volumes, 1)
	require.Contains(t, d.state.volumes, volumeName)
	t.Log("Volume created properly")

	// Verify the service is created
	require.Len(t, d.state.containers, 1)
	require.Contains(t, d.state.containers, containerName)
	container := d.state.containers[containerName]

	// Top-level properties
	require.NotEmpty(t, container.ID)
	require.NotEmpty(t, container.Created)
	require.Equal(t, "/usr/bin/dummy", container.Path)
	require.Equal(t, []string{"arg1", "arg2"}, container.Args)
	require.True(t, container.State.Running)
	require.Equal(t, "/"+containerName, container.Name)
	mounts := []types.MountPoint{
		{
			Type:        mount.TypeBind,
			Source:      "/var/lib/diff-folder",
			Destination: "/diff-folder",
			Mode:        "ro",
			RW:          false,
			Propagation: mount.PropagationRPrivate,
		}, {
			Type:        mount.TypeBind,
			Source:      "/usr/share/same-folder",
			Destination: "/usr/share/same-folder",
			Mode:        "rw",
			RW:          true,
			Propagation: mount.PropagationRPrivate,
		}, {
			Type:        mount.TypeVolume,
			Name:        volumeName,
			Source:      "/var/lib/docker/volumes/" + volumeName + "/_data",
			Destination: "/var/lib/named-vol",
			Driver:      "local",
			Mode:        "z",
			RW:          true,
			Propagation: "",
		},
	}
	require.Equal(t, mounts, container.Mounts)
	t.Log("Service top-level created properly")

	// HostConfig
	require.Equal(t, []string{"/usr/share/same-folder:/usr/share/same-folder:rw", "/var/lib/diff-folder:/diff-folder:ro"}, container.HostConfig.Binds)
	require.Equal(t, dcontainer.NetworkMode(netName), container.HostConfig.NetworkMode)
	portBindings := nat.PortMap{
		"80/tcp": []nat.PortBinding{{HostIP: "127.0.0.1", HostPort: "80"}},
		"81/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "81"}},
		"82/udp": []nat.PortBinding{{HostIP: "", HostPort: "82"}},
	}
	require.Equal(t, portBindings, container.HostConfig.PortBindings)
	require.Equal(t, dcontainer.RestartPolicyUnlessStopped, container.HostConfig.RestartPolicy.Name)
	require.Equal(t, strslice.StrSlice{"dac_override"}, container.HostConfig.CapAdd)
	require.Equal(t, strslice.StrSlice{"all"}, container.HostConfig.CapDrop)
	require.Equal(t, []string{"no-new-privileges"}, container.HostConfig.SecurityOpt)
	hcMounts := []mount.Mount{
		{
			Type:          mount.TypeVolume,
			Source:        volumeName,
			Target:        "/var/lib/named-vol",
			VolumeOptions: &mount.VolumeOptions{},
		},
	}
	require.Equal(t, hcMounts, container.HostConfig.Mounts)
	t.Log("Service HostConfig created properly")

	// NetworkSettings
	nsPorts := nat.PortMap{
		"80/tcp": []nat.PortBinding{{HostIP: "127.0.0.1", HostPort: "80"}},
		"81/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "81"}},
		"82/udp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "82"}, {HostIP: "::", HostPort: "82"}},
	}
	require.Equal(t, nsPorts, container.NetworkSettings.Ports)
	endpointSettings := map[string]*dnetwork.EndpointSettings{
		netName: {
			Aliases:     []string{containerName, goodContainerName},
			MacAddress:  "02:42:ac:12:00:02",
			NetworkID:   network.ID,
			EndpointID:  container.NetworkSettings.Networks[netName].EndpointID,
			Gateway:     "172.18.0.1",
			IPAddress:   "172.18.0.2",
			IPPrefixLen: 16,
			DNSNames:    []string{containerName, goodContainerName, container.ID[:12]},
		},
		"other_net": {
			Aliases:     []string{containerName, goodContainerName},
			MacAddress:  "02:42:ac:11:00:02",
			NetworkID:   d.state.networks["other_net"].ID,
			EndpointID:  container.NetworkSettings.Networks["other_net"].EndpointID,
			Gateway:     "172.17.0.1",
			IPAddress:   "172.17.0.2",
			IPPrefixLen: 16,
			DNSNames:    []string{containerName, goodContainerName, container.ID[:12]},
		},
	}
	require.Equal(t, endpointSettings, container.NetworkSettings.Networks)
	t.Log("Service NetworkSettings created properly")

	// Config
	require.Equal(t, container.ID[:12], container.Config.Hostname)
	require.Equal(t, "dummy-user", container.Config.User)
	exposedPorts := nat.PortSet{
		"80/tcp": struct{}{},
		"81/tcp": struct{}{},
		"82/udp": struct{}{},
	}
	require.Equal(t, exposedPorts, container.Config.ExposedPorts)
	require.Equal(t, []string{"VAR1=VALUE1", "VAR2=VALUE2"}, container.Config.Env)
	require.Equal(t, strslice.StrSlice{"arg1", "arg2"}, container.Config.Cmd)
	require.Equal(t, "mock/test:v0.0.1", container.Config.Image)
	require.Equal(t, strslice.StrSlice{"/usr/bin/dummy"}, container.Config.Entrypoint)
	require.Equal(t, "SIGTERM", container.Config.StopSignal)
	timeout := 300
	require.Equal(t, &timeout, container.Config.StopTimeout)
	t.Log("Service Config created properly")

	// Verify the network has the container in it
	require.Contains(t, network.Containers, container.ID)
	netContainer := network.Containers[container.ID]
	require.Equal(t, containerName, netContainer.Name)
	require.Equal(t, "172.18.0.2", netContainer.IPv4Address)
	require.Equal(t, "02:42:ac:12:00:02", netContainer.MacAddress)
	t.Log("Network has the correct container entry")
}

func TestComposeStop(t *testing.T) {
	// Create a new client
	containerName := fmt.Sprintf("%s_%s", projectName, goodContainerName)
	d := NewDockerMockManager(nil)

	// Run an up command
	err := d.Up(projectName, []string{goodComposeFile})
	if err != nil {
		t.Fatalf("error running compose up: %s", err)
	}
	t.Log("Compose up complete")

	// Run a stop command
	err = d.Stop(projectName, []string{goodComposeFile})
	if err != nil {
		t.Fatalf("error running compose stop: %s", err)
	}
	t.Log("Compose stop complete")

	// Make sure the container is alive but stopped
	require.Contains(t, d.state.containers, containerName)
	container := d.state.containers[containerName]
	require.False(t, container.State.Running)
	t.Log("Container stopped")
}

func TestComposeDown(t *testing.T) {
	// Create a new d
	d := NewDockerMockManager(nil)

	// Run an up command
	err := d.Up(projectName, []string{goodComposeFile})
	if err != nil {
		t.Fatalf("error running compose up: %s", err)
	}
	t.Log("Compose up complete")

	// Make sure the subnets are not available
	require.NotContains(t, d.state.availableSubnets, 17)
	require.NotContains(t, d.state.availableSubnets, 18)
	t.Log("Subnets are not in available pool")

	// Run a stop command
	err = d.Down(projectName, []string{goodComposeFile})
	if err != nil {
		t.Fatalf("error running compose down: %s", err)
	}
	t.Log("Compose down complete")

	// Make sure the container is alive but stopped
	require.Empty(t, d.state.containers)
	require.Empty(t, d.state.volumes)
	require.Empty(t, d.state.networks)
	t.Log("All resources removed")

	// Make sure the subnets are back
	require.Contains(t, d.state.availableSubnets, 17)
	require.Contains(t, d.state.availableSubnets, 18)
	t.Log("Subnets returned to available pool")
}
