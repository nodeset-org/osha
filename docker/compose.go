package docker

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"slices"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/docker/api/types/container"
	"github.com/goccy/go-json"
	"gopkg.in/yaml.v3"
)

// A mock for emulating the Docker compose plugin.
type DockerComposeMock struct {
	// The backing Docker client / database
	client *DockerClientMock

	// A lookup of hashes for services to tell if they need to be regenerated upon starting
	serviceHashes map[string][32]byte
}

// Creates a new compose mock instance
func NewDockerComposeMock(client *DockerClientMock) *DockerComposeMock {
	return &DockerComposeMock{
		client:        client,
		serviceHashes: map[string][32]byte{},
	}
}

// Emulates a `docker compose up` command, starting the provided project by generating the specified
// networks, volumes, and services (containers).
func (d *DockerComposeMock) Up(projectName string, composeFilePaths []string) error {
	// Create the project
	project, err := loadComposeProject(projectName, composeFilePaths)
	if err != nil {
		return fmt.Errorf("error loading project: %w", err)
	}

	// Sort the networks by name
	externalNetworkNames := []string{}
	projectNetworkNames := []string{}
	for nameInYaml, network := range project.Networks {
		if network.External {
			externalNetworkNames = append(externalNetworkNames, nameInYaml)
		} else {
			projectNetworkNames = append(projectNetworkNames, nameInYaml)
		}
	}
	slices.Sort(externalNetworkNames)
	slices.Sort(projectNetworkNames)

	// Create the external networks first
	for _, name := range externalNetworkNames {
		network := project.Networks[name]
		err = d.client.generateNetwork(network, name, projectName)
		if err != nil {
			return err
		}
	}

	// Create the project networks next
	for _, name := range projectNetworkNames {
		network := project.Networks[name]
		err = d.client.generateNetwork(network, name, projectName)
		if err != nil {
			return err
		}
	}

	// Create volumes
	for nameInYaml, volume := range project.Volumes {
		d.client.generateVolume(volume, nameInYaml, projectName)
	}

	// Check each service to see if it needs to be regenerated or just started
	for _, service := range project.Services {
		yamlBytes, err := yaml.Marshal(service)
		if err != nil {
			return fmt.Errorf("error marshalling service [%s]: %w", service.Name, err)
		}

		// Check if the service has changed by hashing the YAML
		hash := sha256.Sum256(yamlBytes)
		existingHash, exists := d.serviceHashes[service.Name]
		if exists && bytes.Equal(hash[:], existingHash[:]) {
			_, exists := d.client.containers[service.Name]
			if exists {
				// Service hasn't changed so just start it
				err = d.client.ContainerStart(context.Background(), service.Name, container.StartOptions{})
				if err != nil {
					return fmt.Errorf("error starting service [%s]: %w", service.Name, err)
				}
				continue
			}
		}

		// Regenerate the service
		err = d.client.generateService(service, project.Networks, project.Volumes)
		if err != nil {
			return err
		}
		d.serviceHashes[service.Name] = hash
	}
	return nil
}

// Emulates a `docker compose stop` command, stopping the provided services.
func (d *DockerComposeMock) Stop(projectName string, composeFilePaths []string) error {
	// Create the project
	project, err := loadComposeProject(projectName, composeFilePaths)
	if err != nil {
		return fmt.Errorf("error loading project: %w", err)
	}

	// Stop each container
	for _, service := range project.Services {
		err = d.client.ContainerStop(context.Background(), service.ContainerName, container.StopOptions{})
		if err != nil {
			return fmt.Errorf("error stopping service [%s]: %w", service.Name, err)
		}
	}
	return nil
}

// Emulates a `docker compose down` command, stopping and removing the provided services, volumes, and networks.
func (d *DockerComposeMock) Down(projectName string, composeFilePaths []string) error {
	// Create the project
	project, err := loadComposeProject(projectName, composeFilePaths)
	if err != nil {
		return fmt.Errorf("error loading project: %w", err)
	}

	// Remove each container
	for _, service := range project.Services {
		err = d.client.ContainerRemove(context.Background(), service.ContainerName, container.RemoveOptions{})
		if err != nil {
			return fmt.Errorf("error removing service [%s]: %w", service.ContainerName, err)
		}
	}

	// Remove each volume
	for _, volume := range project.Volumes {
		err = d.client.VolumeRemove(context.Background(), volume.Name, false)
		if err != nil {
			return fmt.Errorf("error removing volume [%s]: %w", volume.Name, err)
		}
	}

	// Remove each network
	for _, network := range project.Networks {
		err = d.client.NetworkRemove(context.Background(), network.Name)
		if err != nil {
			return fmt.Errorf("error removing network [%s]: %w", network.Name, err)
		}
	}
	return nil
}

// Emulates a `docker compose config` command, returning the YAML configuration for the provided project.
func (d *DockerComposeMock) Config(projectName string, composeFilePaths []string, format ComposeFormat) (string, error) {
	// Create the project options
	options, err := cli.NewProjectOptions(
		composeFilePaths,
		cli.WithName(projectName),
	)
	if err != nil {
		return "", fmt.Errorf("error creating project options: %w", err)
	}

	// Create the project
	model, err := options.LoadModel(context.Background())
	if err != nil {
		return "", fmt.Errorf("error loading model: %w", err)
	}

	// Marshal the model to the requested format
	var modelBytes []byte
	switch format {
	case ComposeFormat_JSON:
		modelBytes, err = json.MarshalIndent(model, "", "  ")
	case ComposeFormat_YAML:
		buf := bytes.NewBuffer([]byte{})
		encoder := yaml.NewEncoder(buf)
		encoder.SetIndent(2)
		err = encoder.Encode(model)
		modelBytes = buf.Bytes()
	}
	if err != nil {
		return "", fmt.Errorf("error marshalling model: %w", err)
	}
	return string(modelBytes), nil
}

// Loads a compose project from the provided compose files
func loadComposeProject(projectName string, composeFiles []string) (*types.Project, error) {
	// Create the project options
	options, err := cli.NewProjectOptions(
		composeFiles,
		cli.WithName(projectName),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating project options: %w", err)
	}

	// Create the project
	return options.LoadProject(context.Background())
}
