package osha

// Interface representing individual module snapshots that compose an entire Snapshot
type IOshaModule interface {
	GetName() string
	Close() error
	TakeSnapshot() (any, error)
	RevertToSnapshot(moduleState any) error
}
