package hyperdrive_constellation

type HyperdriveConstellationModule struct {
	name string
}

func NewHyperdriveConstellationModule() *HyperdriveConstellationModule {
	return &HyperdriveConstellationModule{name: "HyperdriveConstellation"}
}

func (m *HyperdriveConstellationModule) GetName() string {
	return m.name
}

func (m *HyperdriveConstellationModule) GetRequirements() {
}

func (m *HyperdriveConstellationModule) Close() error {
	return nil
}

func (m *HyperdriveConstellationModule) TakeSnapshot() (string, error) {
	return "", nil
}

func (m *HyperdriveConstellationModule) RevertToSnapshot(name string) error {
	return nil
}
