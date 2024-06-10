package docker

// Format for compose output
type ComposeFormat string

const (
	// Output in JSON format
	ComposeFormat_JSON ComposeFormat = "json"

	// Output in YAML format
	ComposeFormat_YAML ComposeFormat = "yaml"
)
