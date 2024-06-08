package osha

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	dclient "github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/nodeset-org/osha/beacon/db"
	"github.com/nodeset-org/osha/beacon/manager"
	"github.com/nodeset-org/osha/docker"
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

	// Docker mock for testing Docker controls
	docker *docker.DockerClientMock

	// Docker compose mock from testing compose functions
	compose *docker.DockerComposeMock

	// Snapshot ID from the baseline - the initial state of Hardhat prior to running any of the tests in this package
	baselineSnapshotID string

	// The Chain ID used by Hardhat
	chainID uint64

	// Path of the temporary directory used for testing
	testDir string
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

	// Create a temp folder
	var err error
	testDir, err := os.MkdirTemp("", "osha-*")
	if err != nil {
		return nil, fmt.Errorf("error creating test dir: %v", err)
	}
	logger.Info("Created test dir", "dir", testDir)

	// Make the RPC client for the Hardhat instance (used for admin functions)
	hardhatRpcClient, err := rpc.Dial(hardhatUrl)
	if err != nil {
		cleanup(testDir)
		return nil, fmt.Errorf("error creating RPC client binding: %w", err)
	}

	// Create a Hardhat client
	primaryEc, err := ethclient.Dial(hardhatUrl)
	if err != nil {
		cleanup(testDir)
		return nil, fmt.Errorf("error creating primary eth client with URL [%s]: %v", hardhatUrl, err)
	}

	// Get the latest block and chain ID from Hardhat
	latestBlockHeader, err := primaryEc.HeaderByNumber(context.Background(), nil)
	if err != nil {
		cleanup(testDir)
		return nil, fmt.Errorf("error getting latest EL block: %v", err)
	}
	chainID, err := primaryEc.ChainID(context.Background())
	if err != nil {
		cleanup(testDir)
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
	docker := docker.NewDockerClientMock()

	m := &TestManager{
		logger:            logger,
		hardhatRpcClient:  hardhatRpcClient,
		executionClient:   primaryEc,
		beaconMockManager: beaconMockManager,
		beaconNode:        beaconNode,
		docker:            docker,
		chainID:           beaconCfg.ChainID,
		testDir:           testDir,
	}

	// Create the baseline snapshot
	baselineSnapshotID, err := m.takeSnapshot()
	if err != nil {
		return nil, fmt.Errorf("error creating baseline snapshot: %w", err)
	}
	m.baselineSnapshotID = baselineSnapshotID

	// Return
	return m, nil
}

// Prints an error message to stderr and exits the program with an error code
func (m *TestManager) Fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
	m.Cleanup()
	os.Exit(1)
}

// Cleans up the test environment, including the testing folder that houses any generated files
func (m *TestManager) Cleanup() {
	err := m.revertToSnapshot(m.baselineSnapshotID)
	if err != nil {
		m.logger.Error("error reverting to baseline snapshot", "err", err)
	}
	if m.testDir != "" {
		cleanup(m.testDir)
	}
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

func (m *TestManager) GetDockerClient() dclient.APIClient {
	return m.docker
}

func (m *TestManager) GetDockerCompose() docker.IDockerCompose {
	return m.compose
}

// Get the path of the test directory - use this to store whatever files you need for testing.
func (m *TestManager) GetTestDir() string {
	return m.testDir
}

// ====================
// === Snapshotting ===
// ====================

// Reverts the EC and BN to the baseline snapshot
func (m *TestManager) RevertToBaseline() error {
	err := m.revertToSnapshot(m.baselineSnapshotID)
	if err != nil {
		return fmt.Errorf("error reverting to baseline snapshot: %w", err)
	}

	// Regenerate the baseline snapshot since Hardhat can't revert to it multiple times
	baselineSnapshotID, err := m.takeSnapshot()
	if err != nil {
		return fmt.Errorf("error creating baseline snapshot: %w", err)
	}
	m.baselineSnapshotID = baselineSnapshotID
	return nil
}

// Takes a snapshot of the EC and BN states
func (m *TestManager) CreateCustomSnapshot() (string, error) {
	return m.takeSnapshot()
}

// Revert the EC and BN to a snapshot state
func (m *TestManager) RevertToCustomSnapshot(snapshotID string) error {
	return m.revertToSnapshot(snapshotID)
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

// ========================
// === Internal Methods ===
// ========================

// Takes a snapshot of the EC and BN states
func (m *TestManager) takeSnapshot() (string, error) {
	// Snapshot the EC
	var snapshotName string
	err := m.hardhatRpcClient.Call(&snapshotName, "evm_snapshot")
	if err != nil {
		return "", fmt.Errorf("error creating snapshot: %w", err)
	}

	// Snapshot the BN
	m.beaconMockManager.TakeSnapshot(snapshotName)
	return snapshotName, nil
}

// Revert the EC and BN to a snapshot state
func (m *TestManager) revertToSnapshot(snapshotID string) error {
	// Revert the EC
	err := m.hardhatRpcClient.Call(nil, "evm_revert", snapshotID)
	if err != nil {
		return fmt.Errorf("error reverting Hardhat to snapshot %s: %w", snapshotID, err)
	}

	// Revert the BN
	err = m.beaconMockManager.RevertToSnapshot(snapshotID)
	if err != nil {
		return fmt.Errorf("error reverting the BN to snapshot %s: %w", snapshotID, err)
	}
	return nil
}

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

// Delete the test dir
func cleanup(testDir string) {
	err := os.RemoveAll(testDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Fprintf(os.Stderr, "error removing test dir [%s]: %v", testDir, err)
	}
}
