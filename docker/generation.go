package docker

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	compose "github.com/compose-spec/compose-go/v2/types"
	docker "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/blkiodev"
	dcontainer "github.com/docker/docker/api/types/container"
	dmount "github.com/docker/docker/api/types/mount"
	dnetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	dvolume "github.com/docker/docker/api/types/volume"
	containerimpl "github.com/docker/docker/container"
	"github.com/docker/go-connections/nat"
	"github.com/docker/go-units"
)

// Generates or updates a network resource from a compose definition.
func (d *DockerMockManager) generateNetwork(network compose.NetworkConfig, nameInYaml string, projectName string) error {
	netResource, exists := d.state.networks[network.Name]
	if exists {
		// Update and return
		netResource.EnableIPv6 = network.EnableIPv6
		netResource.Internal = network.Internal
		netResource.Attachable = network.Attachable
		return nil
	}

	// Get the next available subnet
	if len(d.state.availableSubnets) == 0 {
		return fmt.Errorf("too many networks")
	}
	subnet := d.state.availableSubnets[0]
	d.state.availableSubnets = d.state.availableSubnets[1:]
	d.state.usedSubnets[network.Name] = subnet

	netResource = &docker.NetworkResource{
		Name:       network.Name,
		ID:         createRandomID(32),
		Created:    time.Now(),
		Scope:      "local",
		Driver:     "bridge",
		EnableIPv6: network.EnableIPv6,
		IPAM: dnetwork.IPAM{
			Driver:  "default",
			Options: nil,
			Config: []dnetwork.IPAMConfig{
				{
					Subnet:  fmt.Sprintf("172.%d.0.0/16", subnet),
					Gateway: fmt.Sprintf("172.%d.0.1", subnet),
				},
			},
		},
		Internal:   network.Internal,
		Attachable: network.Attachable,
		Ingress:    false,
		ConfigFrom: dnetwork.ConfigReference{
			Network: "",
		},
		ConfigOnly: false,
		Containers: map[string]docker.EndpointResource{},
		Options:    map[string]string{},
		Labels: map[string]string{
			"com.docker.compose.network": nameInYaml,
			"com.docker.compose.project": projectName,
			"com.docker.compose.version": "0.0.1", // Fake compose version
		},
		Services: map[string]dnetwork.ServiceInfo{}, // Will be filled in later
	}
	d.state.networks[network.Name] = netResource
	d.state.networkIndices[nameInYaml] = 2 // The first IP / MAC for the network starts at 2
	return nil
}

// Generates or updates a volume from a compose definition.
func (d *DockerMockManager) generateVolume(volume compose.VolumeConfig, nameInYaml string, projectName string) {
	volumeImpl, exists := d.state.volumes[volume.Name]
	if exists {
		// Update and return
		volumeImpl.Options = volume.DriverOpts
		return
	}

	volumeImpl = &dvolume.Volume{
		CreatedAt: time.Now().Format(time.RFC3339Nano),
		Driver:    "local",
		Labels: map[string]string{
			"com.docker.compose.project": projectName,
			"com.docker.compose.version": "0.0.1", // Mock version
			"com.docker.compose.volume":  nameInYaml,
		},
		Mountpoint: fmt.Sprintf("/var/lib/docker/volumes/%s/_data", volume.Name),
		Name:       volume.Name,
		Scope:      "local",
		Options:    volume.DriverOpts,
	}
	d.state.volumes[volume.Name] = volumeImpl
}

// Generates a service (container) into the backing database from a compose definition.
// If the service already exists, it will be replaced with a new one.
func (d *DockerMockManager) generateService(service compose.ServiceConfig, projectNetworks compose.Networks, projectVolumes compose.Volumes) error {
	// Create a started stateImpl
	stateImpl := containerimpl.NewState()
	stateImpl.SetRunning(nil, nil, true)
	state := getContainerStateFromState(stateImpl)

	// Create the bind list
	binds := []string{}
	hostConfigMounts := []dmount.Mount{}
	containerMounts := []docker.MountPoint{}
	for _, volume := range service.Volumes {
		switch dmount.Type(volume.Type) {
		case dmount.TypeBind:
			binds = append(binds, volume.String())
			containerMount := docker.MountPoint{
				Type:        dmount.Type(volume.Type),
				Source:      volume.Source,
				Destination: volume.Target,
				Mode:        "rw",
				RW:          !volume.ReadOnly,
				Propagation: dmount.Propagation(volume.Bind.Propagation),
			}
			if !containerMount.RW {
				containerMount.Mode = "ro"
			}
			if containerMount.Propagation == "" {
				containerMount.Propagation = dmount.PropagationRPrivate
			}
			containerMounts = append(containerMounts, containerMount)

		case dmount.TypeVolume:
			projectVolume := projectVolumes[volume.Source]
			volumeResource := d.state.volumes[projectVolume.Name]
			containerMount := docker.MountPoint{
				Type:        dmount.Type(volume.Type),
				Name:        projectVolume.Name,
				Source:      volumeResource.Mountpoint,
				Destination: volume.Target,
				Driver:      "local",
				Mode:        "z",
				RW:          !volume.ReadOnly,
				Propagation: "",
			}
			containerMounts = append(containerMounts, containerMount)

			// Host config mounts seem like they only have volumes
			hostConfigMount := dmount.Mount{
				Type:          containerMount.Type,
				Source:        projectVolume.Name,
				Target:        volume.Target,
				VolumeOptions: &dmount.VolumeOptions{},
			}
			hostConfigMounts = append(hostConfigMounts, hostConfigMount)

		default:
			return fmt.Errorf("unsupported mount type [%s]", volume.Type)
		}
	}
	slices.Sort(binds)

	// Sort ports by string value
	ports := []nat.Port{}
	servicePorts := map[nat.Port]compose.ServicePortConfig{}
	for _, servicePort := range service.Ports {
		target := strconv.FormatUint(uint64(servicePort.Target), 10)
		port, err := nat.NewPort(servicePort.Protocol, target)
		if err != nil {
			return fmt.Errorf("error parsing service port binding [%s] on container [%s]: %w", target, service.Name, err)
		}
		ports = append(ports, port)
		servicePorts[port] = servicePort
	}
	slices.Sort(ports)

	// Make the port maps
	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}
	nsPorts := nat.PortMap{}
	for _, port := range ports {
		servicePort := servicePorts[port]

		// HostConfig port bindings
		portBindings[port] = []nat.PortBinding{
			{
				HostIP:   servicePort.HostIP,
				HostPort: servicePort.Published,
			},
		}

		// Config exposed ports
		exposedPorts[port] = struct{}{}

		// NetworkSettings ports
		if servicePort.HostIP == "" {
			nsPorts[port] = []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: servicePort.Published,
				}, {
					HostIP:   "::",
					HostPort: servicePort.Published,
				},
			}
		} else {
			nsPorts[port] = []nat.PortBinding{
				{
					HostIP:   servicePort.HostIP,
					HostPort: servicePort.Published,
				},
			}
		}
	}

	// Make the environment variable list
	env := []string{}
	for key, val := range service.Environment {
		env = append(env, fmt.Sprintf("%s=%s", key, *val))
	}
	slices.Sort(env)

	// Make the network list
	networks := map[string]*dnetwork.EndpointSettings{}
	for networkNameInYaml, networkConfig := range service.Networks {
		// Get the corresponding network
		projectNetwork := projectNetworks[networkNameInYaml]
		netResource := d.state.networks[projectNetwork.Name]
		subnet := d.state.usedSubnets[projectNetwork.Name]
		indexInNetwork := d.state.networkIndices[networkNameInYaml]

		// Set up the compose file config if it's not defined
		if networkConfig == nil {
			networkConfig = &compose.ServiceNetworkConfig{
				Aliases: []string{
					service.ContainerName,
					service.Name,
				},
				MacAddress: fmt.Sprintf("02:42:ac:%02x:00:%02x", subnet, indexInNetwork),
			}
		}

		network := &dnetwork.EndpointSettings{
			Aliases:     networkConfig.Aliases,
			MacAddress:  networkConfig.MacAddress,
			NetworkID:   netResource.ID,
			EndpointID:  createRandomID(32),
			Gateway:     netResource.IPAM.Config[0].Gateway,
			IPAddress:   fmt.Sprintf("172.%d.0.%d", subnet, indexInNetwork),
			IPPrefixLen: 16,
			DNSNames:    []string{
				// This will be filled in at the end since it also requires the service ID
			},
		}
		networks[projectNetwork.Name] = network

		// Increment the network index for the next service using it
		d.state.networkIndices[networkNameInYaml] = indexInNetwork + 1
	}

	// Create the container
	id := createRandomID(32)
	image := createRandomID(32)
	networkSandboxID := createRandomID(32)
	container := &docker.ContainerJSON{
		ContainerJSONBase: &docker.ContainerJSONBase{
			ID:              id,
			Created:         time.Now().Format(time.RFC3339Nano),
			Path:            strings.Join(service.Entrypoint, " "),
			Args:            service.Command,
			State:           &state,
			Image:           fmt.Sprintf("sha256:%s", image),
			ResolvConfPath:  fmt.Sprintf("/var/lib/docker/containers/%s/resolv.conf", id),
			HostnamePath:    fmt.Sprintf("/var/lib/docker/containers/%s/hostname", id),
			HostsPath:       fmt.Sprintf("/var/lib/docker/containers/%s/hosts", id),
			LogPath:         fmt.Sprintf("/var/lib/docker/containers/%s/%s-json.log", id, id),
			Name:            "/" + service.ContainerName,
			RestartCount:    0,
			Driver:          "overlay2",
			Platform:        "linux",
			MountLabel:      "",
			ProcessLabel:    "",
			AppArmorProfile: "",
			ExecIDs:         nil,
			HostConfig: &dcontainer.HostConfig{
				Binds:           binds,
				ContainerIDFile: "",
				LogConfig: dcontainer.LogConfig{
					Type:   "json-file",
					Config: map[string]string{},
				},
				NetworkMode:  dcontainer.NetworkMode(service.NetworkMode),
				PortBindings: portBindings,
				RestartPolicy: dcontainer.RestartPolicy{
					Name:              dcontainer.RestartPolicyMode(service.Restart),
					MaximumRetryCount: 0,
				},
				AutoRemove:      false,
				VolumeDriver:    service.VolumeDriver,
				VolumesFrom:     service.VolumesFrom,
				ConsoleSize:     [2]uint{0, 0},
				Annotations:     service.Annotations,
				CapAdd:          service.CapAdd,
				CapDrop:         service.CapDrop,
				CgroupnsMode:    dcontainer.CgroupnsModePrivate,
				DNS:             service.DNS,
				DNSOptions:      service.DNSOpts,
				DNSSearch:       service.DNSSearch,
				ExtraHosts:      nil,
				GroupAdd:        service.GroupAdd,
				IpcMode:         dcontainer.IPCModePrivate,
				Cgroup:          dcontainer.CgroupSpec(service.Cgroup),
				Links:           service.Links,
				OomScoreAdj:     int(service.OomScoreAdj),
				PidMode:         "",
				Privileged:      service.Privileged,
				PublishAllPorts: false,
				ReadonlyRootfs:  service.ReadOnly,
				SecurityOpt:     service.SecurityOpt,
				StorageOpt:      service.StorageOpt,
				Tmpfs:           nil, //service.Tmpfs // TODO: figure out what the keys should be here
				UTSMode:         dcontainer.UTSMode(service.Uts),
				UsernsMode:      dcontainer.UsernsMode(service.UserNSMode),
				ShmSize:         int64(service.ShmSize),
				Sysctls:         service.Sysctls,
				Runtime:         service.Runtime,
				Isolation:       dcontainer.Isolation(service.Isolation),
				Resources: dcontainer.Resources{
					CPUShares:          service.CPUShares,
					Memory:             int64(service.MemLimit),
					CgroupParent:       service.CgroupParent,
					CPUPeriod:          service.CPUPeriod,
					CPUQuota:           service.CPUQuota,
					CPURealtimePeriod:  service.CPURTPeriod,
					CPURealtimeRuntime: service.CPURTRuntime,
					CpusetCpus:         service.CPUSet,
					CpusetMems:         service.CPUSet, // ?
					Devices:            nil,            //service.Devices // TODO: figure out the mapping here
					DeviceCgroupRules:  service.DeviceCgroupRules,
					DeviceRequests:     nil, // ?
					MemoryReservation:  int64(service.MemReservation),
					MemorySwap:         int64(service.MemSwapLimit),
					MemorySwappiness:   nil,
					OomKillDisable:     &service.OomKillDisable,
					PidsLimit:          &service.PidsLimit,
					Ulimits:            nil,
					CPUCount:           service.CPUCount,
					CPUPercent:         int64(service.CPUPercent * 100), // ?
				},
				Mounts: hostConfigMounts,
				MaskedPaths: []string{
					"/proc/asound",
					"/proc/acpi",
					"/proc/kcore",
					"/proc/keys",
					"/proc/latency_stats",
					"/proc/timer_list",
					"/proc/timer_stats",
					"/proc/sched_debug",
					"/proc/scsi",
					"/sys/firmware",
					"/sys/devices/virtual/powercap",
				},
				ReadonlyPaths: []string{
					"/proc/bus",
					"/proc/fs",
					"/proc/irq",
					"/proc/sys",
					"/proc/sysrq-trigger",
				},
			},
			GraphDriver: docker.GraphDriverData{}, // Not implemented
			SizeRw:      new(int64),
			SizeRootFs:  new(int64),
		},
		Mounts: containerMounts,
		NetworkSettings: &docker.NetworkSettings{
			NetworkSettingsBase: docker.NetworkSettingsBase{
				Bridge:                 "",
				SandboxID:              networkSandboxID,
				SandboxKey:             fmt.Sprintf("/var/run/docker/netns/%s", networkSandboxID[:12]),
				Ports:                  nsPorts,
				LinkLocalIPv6Address:   "", // ?
				LinkLocalIPv6PrefixLen: 0,  // ?
			},
			Networks: networks,
		},
		Config: &dcontainer.Config{
			Hostname:     service.Hostname,
			Domainname:   service.DomainName,
			User:         service.User,
			AttachStdin:  false,
			AttachStdout: true,
			AttachStderr: true,
			ExposedPorts: exposedPorts, // NOTE: this should include ports exposed by the image itself too, but it won't because we don't actually pull the image
			Tty:          service.Tty,
			OpenStdin:    service.StdinOpen,
			StdinOnce:    false, // ?
			Env:          env,   // NOTE: this should include env vars set by the image itself too, but it won't because we don't actually pull the image
			Cmd:          strslice.StrSlice(service.Command),
			Image:        service.Image,
			Volumes:      map[string]struct{}{},                 // This doesn't look like it uses the actual volume list from the service?
			WorkingDir:   service.WorkingDir,                    // Should come from the image if not set
			Entrypoint:   strslice.StrSlice(service.Entrypoint), // Should come from the image if not set
			MacAddress:   service.MacAddress,
			StopSignal:   service.StopSignal,
		},
	}

	// Populate fields that get set at runtime normally instead of via the compose files
	if service.NetworkMode == "" {
		for netNameInYaml := range service.Networks {
			network := projectNetworks[netNameInYaml]
			if network.External {
				continue
			}
			container.HostConfig.NetworkMode = dcontainer.NetworkMode(network.Name)
			break
		}
	}
	if service.Runtime == "" {
		container.HostConfig.Runtime = "runc"
	}
	if service.MemSwappiness > 0 {
		swappiness := int64(service.MemSwappiness)
		container.HostConfig.MemorySwappiness = &swappiness
	}
	if service.Hostname == "" {
		container.Config.Hostname = container.ID[:12] // First 12 chars of ID
	}
	for _, network := range container.NetworkSettings.Networks {
		network.DNSNames = append(network.DNSNames, network.Aliases...)
		network.DNSNames = append(network.DNSNames, container.Config.Hostname)
	}

	// Populate fields that need nil checks
	if service.ExtraHosts != nil {
		container.HostConfig.ExtraHosts = service.ExtraHosts.AsList(":")
	}
	if service.HealthCheck != nil {
		container.HostConfig.RestartPolicy.MaximumRetryCount = int(*service.HealthCheck.Retries)
	}
	if service.BlkioConfig != nil {
		container.HostConfig.Resources.BlkioWeight = service.BlkioConfig.Weight
		for _, element := range service.BlkioConfig.WeightDevice {
			container.HostConfig.Resources.BlkioWeightDevice = append(container.HostConfig.Resources.BlkioWeightDevice, &blkiodev.WeightDevice{
				Path:   element.Path,
				Weight: element.Weight,
			})
		}
		for _, element := range service.BlkioConfig.DeviceReadBps {
			container.HostConfig.Resources.BlkioDeviceReadBps = append(container.HostConfig.Resources.BlkioDeviceReadBps, &blkiodev.ThrottleDevice{
				Path: element.Path,
				Rate: uint64(element.Rate),
			})
		}
		for _, element := range service.BlkioConfig.DeviceReadIOps {
			container.HostConfig.Resources.BlkioDeviceReadIOps = append(container.HostConfig.Resources.BlkioDeviceReadIOps, &blkiodev.ThrottleDevice{
				Path: element.Path,
				Rate: uint64(element.Rate),
			})
		}
		for _, element := range service.BlkioConfig.DeviceWriteBps {
			container.HostConfig.Resources.BlkioDeviceWriteBps = append(container.HostConfig.Resources.BlkioDeviceWriteBps, &blkiodev.ThrottleDevice{
				Path: element.Path,
				Rate: uint64(element.Rate),
			})
		}
		for _, element := range service.BlkioConfig.DeviceWriteIOps {
			container.HostConfig.Resources.BlkioDeviceWriteIOps = append(container.HostConfig.Resources.BlkioDeviceWriteIOps, &blkiodev.ThrottleDevice{
				Path: element.Path,
				Rate: uint64(element.Rate),
			})
		}
	}
	if service.Ulimits != nil {
		for name, limit := range service.Ulimits {
			container.HostConfig.Resources.Ulimits = append(container.HostConfig.Resources.Ulimits, &units.Ulimit{
				Name: name,
				Hard: int64(limit.Hard),
				Soft: int64(limit.Soft),
			})
		}
	}
	if service.StopGracePeriod != nil {
		timeout := int(time.Duration(*service.StopGracePeriod).Seconds())
		container.Config.StopTimeout = &timeout
	}

	d.state.containers[service.ContainerName] = container

	// Add the container to the networks
	for netName, network := range container.NetworkSettings.Networks {
		netResource := d.state.networks[netName]
		netResource.Containers[container.ID] = docker.EndpointResource{
			Name:        strings.TrimPrefix(container.Name, "/"),
			EndpointID:  network.EndpointID,
			MacAddress:  network.MacAddress,
			IPv4Address: network.IPAddress,
		}
	}
	return nil
}

// Creates a random hex-encoded ID string
func createRandomID(byteCount int) string {
	// Create a fake ID
	bytes := make([]byte, byteCount)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
