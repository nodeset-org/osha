package osha

// Interface representing individual module snapshots that compose an entire Snapshot
type IOshaModule interface {
	GetModuleName() string
	CloseModule() error
	TakeModuleSnapshot() (any, error)
	RevertModuleToSnapshot(moduleState any) error
}
