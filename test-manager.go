package osha

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/google/uuid"
	"github.com/nodeset-org/osha/beacon/db"
	"github.com/nodeset-org/osha/beacon/manager"
	"github.com/nodeset-org/osha/docker"
	"github.com/nodeset-org/osha/filesystem"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/beacon/client"
	"github.com/rocket-pool/node-manager-core/eth"
)

const (
	// The environment variable for the locally running Hardhat instance
	HardhatEnvVar string = "HARDHAT_URL"
)

// TestManager provides bootstrapping and a test service provider, useful for testing
type TestManager struct {
	// logger for logging output messages during tests
	logger *slog.Logger

	// RPC client for running Hardhat's admin functions
	hardhatRpcClient *rpc.Client

	// Execution client for Hardhat's ETH API
	executionClient eth.IExecutionClient

	// Beacon mock manager for running BN admin functions
	beaconMockManager *manager.BeaconMockManager

	// Beacon node
	beaconNode beacon.IBeaconClient

	// Docker mock for testing Docker controls and compose functions
	docker *docker.DockerMockManager

	// Snapshot ID from the baseline - the initial state of Hardhat prior to running any of the tests in this package
	baselineSnapshotID string

	// The Chain ID used by Hardhat
	chainID uint64

	// Manager for the filesystem's test folder
	fsManager *filesystem.FilesystemManager

	// Map of snapshot name to snapshot for registered modules (unique UUID => snapshot)
	snapshots map[string]Snapshot

	// Map of OSHA snapshot name mapped to hardhat snapshot name (unique UUID => hardhat-specific ID)
	hardhatSnapshotMap map[string]string

	// Map of registered modules (moduleName -> module)
	registeredModules map[string]IOshaModule
}

// Creates a new TestManager instance
func NewTestManager() (*TestManager, error) {
	// Make sure the Hardhat URL
	hardhatUrl, exists := os.LookupEnv(HardhatEnvVar)
	if !exists {
		return nil, fmt.Errorf("%s env var not set", HardhatEnvVar)
	}

	// Make a new logger
	logger := slog.Default()

	// Make the FS manager
	fsManager, err := filesystem.NewFilesystemManager(logger)
	if err != nil {
		return nil, fmt.Errorf("error creating FS manager: %w", err)
	}

	// Make the RPC client for the Hardhat instance (used for admin functions)
	hardhatRpcClient, err := rpc.Dial(hardhatUrl)
	if err != nil {
		err2 := fsManager.Close()
		if err2 != nil {
			logger.Error("error closing FS manager", "err", err)
		}
		return nil, fmt.Errorf("error creating RPC client binding: %w", err)
	}

	// Create a Hardhat client
	primaryEc, err := ethclient.Dial(hardhatUrl)
	if err != nil {
		err2 := fsManager.Close()
		if err2 != nil {
			logger.Error("error closing FS manager", "err", err)
		}
		return nil, fmt.Errorf("error creating primary eth client with URL [%s]: %v", hardhatUrl, err)
	}

	// Get the latest block and chain ID from Hardhat
	latestBlockHeader, err := primaryEc.HeaderByNumber(context.Background(), nil)
	if err != nil {
		err2 := fsManager.Close()
		if err2 != nil {
			logger.Error("error closing FS manager", "err", err)
		}
		return nil, fmt.Errorf("error getting latest EL block: %v", err)
	}
	chainID, err := primaryEc.ChainID(context.Background())
	if err != nil {
		err2 := fsManager.Close()
		if err2 != nil {
			logger.Error("error closing FS manager", "err", err)
		}
		return nil, fmt.Errorf("error getting chain ID: %v", err)
	}

	// Create the Beacon config based on the Hardhat values
	beaconCfg := db.NewDefaultConfig()
	beaconCfg.FirstExecutionBlockIndex = latestBlockHeader.Number.Uint64()
	beaconCfg.ChainID = chainID.Uint64()
	beaconCfg.GenesisTime = time.Unix(int64(latestBlockHeader.Time), 0)

	// Make the Beacon client manager
	beaconMockManager := manager.NewBeaconMockManager(logger, beaconCfg)
	beaconNode := client.NewStandardClient(beaconMockManager)

	// Make a Docker client mock
	docker := docker.NewDockerMockManager(logger)

	m := &TestManager{
		logger:             logger,
		hardhatRpcClient:   hardhatRpcClient,
		executionClient:    primaryEc,
		beaconMockManager:  beaconMockManager,
		beaconNode:         beaconNode,
		docker:             docker,
		chainID:            beaconCfg.ChainID,
		fsManager:          fsManager,
		snapshots:          map[string]Snapshot{},
		hardhatSnapshotMap: map[string]string{},
		registeredModules:  map[string]IOshaModule{},
	}

	// Create the baseline snapshot
	baselineSnapshotID, err := m.CreateSnapshot()
	if err != nil {
		return nil, fmt.Errorf("error creating baseline snapshot: %w", err)
	}
	m.baselineSnapshotID = baselineSnapshotID

	// Return
	return m, nil
}

// Manages test dependencies for running individual unit tests when previous snapshots are not available
func (m *TestManager) DependsOn(dependency func(*testing.T), snapshotName *string, t *testing.T) error {
	if snapshotName != nil && *snapshotName != "" {
		err := m.RevertSnapshot(*snapshotName)
		if err != nil {
			return fmt.Errorf("error reverting to snapshot %s: %v", *snapshotName, err)
		}
		return nil
	}
	dependency(t)
	return nil
}

// Cleans up the test environment, including the testing folder that houses any generated files
func (m *TestManager) Close() error {
	err := m.RevertSnapshot(m.baselineSnapshotID)
	if err != nil {
		return fmt.Errorf("error reverting to baseline snapshot: %w", err)
	}
	return m.fsManager.Close()
}

// ===============
// === Getters ===
// ===============

func (m *TestManager) GetLogger() *slog.Logger {
	return m.logger
}

func (m *TestManager) GetHardhatRpcClient() *rpc.Client {
	return m.hardhatRpcClient
}

func (m *TestManager) GetExecutionClient() eth.IExecutionClient {
	return m.executionClient
}

func (m *TestManager) GetBeaconMockManager() *manager.BeaconMockManager {
	return m.beaconMockManager
}

func (m *TestManager) GetBeaconClient() beacon.IBeaconClient {
	return m.beaconNode
}

func (m *TestManager) GetDockerMockManager() *docker.DockerMockManager {
	return m.docker
}

// Get the path of the test directory - use this to store whatever files you need for testing.
func (m *TestManager) GetTestDir() string {
	return m.fsManager.GetTestDir()
}

// ====================
// === Snapshotting ===
// ====================

// Reverts the services to the baseline snapshot
func (m *TestManager) RevertToBaseline() error {
	err := m.RevertSnapshot(m.baselineSnapshotID)
	if err != nil {
		return fmt.Errorf("error reverting to baseline snapshot: %w", err)
	}

	return nil
}

// Takes a snapshot of the service states
func (m *TestManager) CreateSnapshot() (string, error) {
	var snapshotName string
	for {
		candidateName := uuid.New().String()
		_, exists := m.snapshots[candidateName]
		if !exists {
			snapshotName = candidateName
			break
		}
	}

	// Create a new snapshot
	snapshot := Snapshot{
		name:   snapshotName,
		states: make(map[IOshaModule]any),
	}
	var hardhatSnapshotName string
	// Take a snapshot of hardhat
	err := m.hardhatRpcClient.Call(&hardhatSnapshotName, "evm_snapshot")
	if err != nil {
		return "", fmt.Errorf("error taking snapshot of Hardhat: %w", err)
	}
	m.hardhatSnapshotMap[snapshotName] = hardhatSnapshotName

	// Take a snapshot of the BN
	m.beaconMockManager.TakeSnapshot(snapshotName)

	// Take a snapshot of Docker
	err = m.docker.TakeSnapshot(snapshotName)
	if err != nil {
		return "", fmt.Errorf("error taking snapshot of Docker: %w", err)
	}

	// Take a snapshot of the filesystem
	err = m.fsManager.TakeSnapshot(snapshotName)
	if err != nil {
		return "", fmt.Errorf("error taking snapshot of the filesystem: %w", err)
	}

	// Take a snapshot of all registered modules
	for _, module := range m.registeredModules {
		state, err := module.TakeModuleSnapshot()
		if err != nil {
			return "", fmt.Errorf("error taking snapshot for module %s: %w", module.GetModuleName(), err)
		}
		snapshot.states[module] = state
	}

	// Store the snapshot
	m.snapshots[snapshotName] = snapshot

	return snapshotName, nil
}

func (m *TestManager) RevertSnapshot(snapshotName string) error {
	snapshot, exists := m.snapshots[snapshotName]
	if !exists {
		return fmt.Errorf("snapshot %s does not exist", snapshotName)
	}

	// Revert snapshot of Hardhat
	hardhatSnapshotName, exists := m.hardhatSnapshotMap[snapshotName]
	if !exists {
		return fmt.Errorf("Hardhat snapshot ID not found for snapshot ID [%s]", snapshotName)
	}
	err := m.hardhatRpcClient.Call(nil, "evm_revert", hardhatSnapshotName)
	if err != nil {
		return fmt.Errorf("error reverting Hardhat to snapshot %s: %w", snapshotName, err)
	}

	// Take a snapshot of Hardhat again because hardhat deletes reverted snapshots
	err = m.hardhatRpcClient.Call(&hardhatSnapshotName, "evm_snapshot")
	if err != nil {
		return fmt.Errorf("error regenerating snapshot of Hardhat after revert: Hardhat deletes reverted snapshots. %w", err)
	}
	m.hardhatSnapshotMap[snapshotName] = hardhatSnapshotName

	// Revert the BN
	err = m.beaconMockManager.RevertToSnapshot(snapshotName)
	if err != nil {
		return fmt.Errorf("error reverting the BN to snapshot %s: %w", snapshotName, err)
	}

	// Revert Docker
	err = m.docker.RevertToSnapshot(snapshotName)
	if err != nil {
		return fmt.Errorf("error reverting Docker to snapshot %s: %w", snapshotName, err)
	}

	// Revert the filesystem
	err = m.fsManager.RevertToSnapshot(snapshotName)
	if err != nil {
		return fmt.Errorf("error reverting the filesystem to snapshot %s: %w", snapshotName, err)
	}

	// Revert all registered modules
	for _, module := range m.registeredModules {
		moduleState, exists := snapshot.states[module]
		if !exists {
			continue
		}
		err := module.RevertModuleToSnapshot(moduleState)
		if err != nil {
			return fmt.Errorf("error reverting the module to snapshot %s: %w", moduleState, err)
		}
	}

	return nil
}

// If a user registers a module with an existing name, it will be overwritten
func (m *TestManager) RegisterModule(module IOshaModule) {
	m.registeredModules[module.GetModuleName()] = module
}

// Returns a list of registered modules
func (m *TestManager) GetRegisteredModules() []IOshaModule {
	modules := make([]IOshaModule, 0, len(m.registeredModules))
	for _, module := range m.registeredModules {
		modules = append(modules, module)
	}
	return modules
}

// ==========================
// === Chain Modification ===
// ==========================

// Commits a new block in the EC and BN, advancing the chain
func (m *TestManager) CommitBlock() error {
	// Mine the next block in Hardhat
	err := m.hardhat_mineBlock()
	if err != nil {
		return err
	}

	// Increase time by the slot duration to prep for the next slot
	secondsPerSlot := uint(m.beaconMockManager.GetConfig().SecondsPerSlot)
	err = m.hardhat_increaseTime(secondsPerSlot)
	if err != nil {
		return err
	}

	// Commit the block in the BN
	m.beaconMockManager.CommitBlock(true)
	return nil
}

// Advances the chain by a number of slots.
// If includeBlocks is true, an EL block will be mined for each slot and the slot will reference that block.
// If includeBlocks is false, each slot (until the last one) will be "missed", so no EL block will be mined for it.
func (m *TestManager) AdvanceSlots(slots uint, includeBlocks bool) error {
	if includeBlocks {
		for i := uint(0); i < slots; i++ {
			err := m.CommitBlock()
			if err != nil {
				return err
			}
		}
		return nil
	}

	// Commit slots without blocks
	for i := uint(0); i < slots; i++ {
		m.beaconMockManager.CommitBlock(false)
	}

	// Advance the time in Hardhat
	secondsPerSlot := uint(m.beaconMockManager.GetConfig().SecondsPerSlot)
	err := m.hardhatRpcClient.Call(nil, "evm_increaseTime", secondsPerSlot*slots)
	if err != nil {
		return fmt.Errorf("error advancing time on EL: %w", err)
	}
	return nil
}

// Set the highest slot (the head slot) of the Beacon chain, while keeping the local chain head on the client the same.
// Useful for simulating an unsynced client.
func (m *TestManager) SetBeaconHeadSlot(slot uint64) {
	m.beaconMockManager.SetHighestSlot(slot)
}

// Toggle automining where each TX will automatically be mine into its own block
func (m *TestManager) ToggleAutoMine(enabled bool) error {
	err := m.hardhatRpcClient.Call(nil, "evm_setAutomine", enabled)
	if err != nil {
		return fmt.Errorf("error toggling automine: %w", err)
	}
	return nil
}

// Set the interval for interval mining mode
func (m *TestManager) SetMiningInterval(interval uint) error {
	err := m.hardhatRpcClient.Call(nil, "evm_setIntervalMining", interval)
	if err != nil {
		return fmt.Errorf("error setting interval mining: %w", err)
	}
	return nil
}

// ========================
// === Internal Methods ===
// ========================

// Tell Hardhat to mine a block
func (m *TestManager) hardhat_mineBlock() error {
	err := m.hardhatRpcClient.Call(nil, "evm_mine")
	if err != nil {
		return fmt.Errorf("error mining EL block: %w", err)
	}
	return nil
}

// Tell Hardhat to mine a block
func (m *TestManager) hardhat_increaseTime(seconds uint) error {
	err := m.hardhatRpcClient.Call(nil, "evm_increaseTime", seconds)
	if err != nil {
		return fmt.Errorf("error increasing EL time: %w", err)
	}
	return nil
}
