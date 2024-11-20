package osha

// Interface representing individual module snapshots that compose an entire Snapshot
type IOshaModule interface {
	GetName() string
	GetRequirements()
	Close() error
	TakeSnapshot() (any, error)
	RevertToSnapshot(name any) error
}
