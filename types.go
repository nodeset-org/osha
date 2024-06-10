package osha

// Service represents a service provided by OSHA
type Service int

const (
	// Represents the Execution client and Beacon node services
	Service_EthClients Service = 1 << iota

	// Represents the Docker client and compose services
	Service_Docker

	// Represents the underlying filesystem - NYI
	//Service_Filesystem Service = "filesystem"

	// Represents all of the services provided by OSHA
	Service_All Service = Service_EthClients | Service_Docker
)

// Check if a service value contains a specific service flag
func (s Service) Contains(service Service) bool {
	return s&service == service
}
