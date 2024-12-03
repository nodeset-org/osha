package osha

// Struct representing an entire snapshot for a given test case
type Snapshot struct {
	name   string
	states map[IOshaModule]string
}
