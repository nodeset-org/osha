package docker

import (
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/goccy/go-json"
)

// Underlying state for the Docker mock
type state struct {
	// Mock database fields
	containers map[string]*types.ContainerJSON
	volumes    map[string]*volume.Volume
	networks   map[string]*network.Inspect

	// Internal fields
	availableSubnets []int
	usedSubnets      map[string]int

	// Used to create sequential IP / MAC addresses for services
	networkIndices map[string]byte

	// A lookup of hashes for services to tell if they need to be regenerated upon starting
	serviceHashes map[string][32]byte
}

// Creates a new Docker state
func newState() *state {
	// Docker defaults to the available subnets for bridges: 172.17.0.0/16 through 172.31.0.0/16
	availableSubnets := []int{}
	for i := 17; i < 32; i++ {
		availableSubnets = append(availableSubnets, i)
	}

	return &state{
		containers:       map[string]*types.ContainerJSON{},
		volumes:          map[string]*volume.Volume{},
		networks:         map[string]*network.Inspect{},
		availableSubnets: availableSubnets,
		usedSubnets:      map[string]int{},
		networkIndices:   map[string]byte{},
		serviceHashes:    map[string][32]byte{},
	}
}

// Clone the current state
func (s *state) Clone() (*state, error) {
	clone := newState()

	// Since all of these types are supposed to be JSON-exportable, it's easiest to just serialize and deserialize them

	// Copy the containers
	for name, origContainer := range s.containers {
		// Serialize the original
		marshalled, err := json.Marshal(origContainer)
		if err != nil {
			return nil, fmt.Errorf("error marshalling container [%s]: %w", name, err)
		}

		// Deserialize into a clone
		var cloneContainer types.ContainerJSON
		err = json.Unmarshal(marshalled, &cloneContainer)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling container [%s]: %w", name, err)
		}
		clone.containers[name] = &cloneContainer
	}

	// Copy the volumes
	for name, origVolume := range s.volumes {
		// Serialize the original
		marshalled, err := json.Marshal(origVolume)
		if err != nil {
			return nil, fmt.Errorf("error marshalling volume [%s]: %w", name, err)
		}

		// Deserialize into a clone
		var cloneVolume volume.Volume
		err = json.Unmarshal(marshalled, &cloneVolume)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling volume [%s]: %w", name, err)
		}
		clone.volumes[name] = &cloneVolume
	}

	// Copy the networks
	for name, origNetwork := range s.networks {
		// Serialize the original
		marshalled, err := json.Marshal(origNetwork)
		if err != nil {
			return nil, fmt.Errorf("error marshalling network [%s]: %w", name, err)
		}

		// Deserialize into a clone
		var cloneNetwork network.Inspect
		err = json.Unmarshal(marshalled, &cloneNetwork)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling network [%s]: %w", name, err)
		}
		clone.networks[name] = &cloneNetwork
	}

	// Copy the available subnets
	clone.availableSubnets = make([]int, len(s.availableSubnets))
	copy(clone.availableSubnets, s.availableSubnets)

	// Copy the used subnets
	clone.usedSubnets = map[string]int{}
	for subnet, index := range s.usedSubnets {
		clone.usedSubnets[subnet] = index
	}

	// Copy the network indices
	clone.networkIndices = map[string]byte{}
	for network, index := range s.networkIndices {
		clone.networkIndices[network] = index
	}

	return clone, nil
}
