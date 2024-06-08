package docker

// Format for compose output
type ComposeFormat string

const (
	// Output in JSON format
	ComposeFormat_JSON ComposeFormat = "json"

	// Output in YAML format
	ComposeFormat_YAML ComposeFormat = "yaml"
)

// Interface for a Docker Compose client
type IDockerCompose interface {
	// Runs the `docker compose up` command, instantiating a project from its compose files
	Up(projectName string, composeFilePaths []string) error

	// Runs the `docker compose stop` command, shutting down running containers without deleting them
	Stop(projectName string, composeFilePaths []string) error

	// Runs the `docker compose down` command, shutting down all of the services in the project
	// and deleting them, along with the project's networks and volumes
	Down(projectName string, composeFilePaths []string) error

	// Runs the `docker compose config` command, returning the complete configuration for the provided project
	// across all of its compose files without instantiating it
	Config(projectName string, composeFilePaths []string, format ComposeFormat) (string, error)
}
