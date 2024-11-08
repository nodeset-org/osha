package osha

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/google/uuid"
	beacondb "github.com/nodeset-org/osha/beacon/db"
	bnmanager "github.com/nodeset-org/osha/beacon/manager"
	bnserver "github.com/nodeset-org/osha/beacon/server"
	"github.com/nodeset-org/osha/docker"
	"github.com/nodeset-org/osha/filesystem"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/beacon/client"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/log"
)

const (
	// The environment variable for the locally running Hardhat instance
	HardhatEnvVar string = "HARDHAT_URL"
)

type TestManagerOptions struct {
	// The URL of the locally running Hardhat instance
	HardhatUrl *string

	// The logger for logging output messages during tests
	Logger *slog.Logger

	// The Beacon Client configuration
	BeaconConfig *beacondb.Config

	// The hostname to run the Beacon Node with
	BeaconHostname *string

	// The port to run the Beacon Node with
	BeaconPort *uint16
}

// TestManager provides bootstrapping and a test service provider, useful for testing
type TestManager struct {
	// logger for logging output messages during tests
	logger *slog.Logger

	// RPC client for running Hardhat's admin functions
	hardhatRpcClient *rpc.Client

	// Execution client for Hardhat's ETH API
	executionClient eth.IExecutionClient

	// Beacon mock server for running BN admin functions
	beaconMockServer *bnserver.BeaconMockServer

	// Beacon node
	beaconNode beacon.IBeaconClient

	// BN Wait group for graceful shutdown
	bnWg *sync.WaitGroup

	// Docker mock for testing Docker controls and compose functions
	docker *docker.DockerMockManager

	// Snapshot ID from the baseline - the initial state of Hardhat prior to running any of the tests in this package
	baselineSnapshotID string

	// The Chain ID used by Hardhat
	chainID uint64

	// Manager for the filesystem's test folder
	fsManager *filesystem.FilesystemManager

	// Map of which services were captured during a snapshot
	snapshotServiceMap map[string]Service
}

// Creates a new TestManager instance
func NewTestManager(opts *TestManagerOptions) (*TestManager, error) {
	// Set up the options
	if opts == nil {
		opts = &TestManagerOptions{}
	}
	if opts.HardhatUrl == nil {
		hardhatUrl, exists := os.LookupEnv(HardhatEnvVar)
		if !exists {
			return nil, fmt.Errorf("%s env var not set", HardhatEnvVar)
		}
		opts.HardhatUrl = &hardhatUrl
	}
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	if opts.BeaconConfig == nil {
		opts.BeaconConfig = beacondb.NewDefaultConfig()
	}
	if opts.BeaconHostname == nil {
		hostname := "localhost"
		opts.BeaconHostname = &hostname
	}
	if opts.BeaconPort == nil {
		port := uint16(5052)
		opts.BeaconPort = &port
	}

	// Load the options
	hardhatUrl := *opts.HardhatUrl
	logger := opts.Logger

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
	beaconCfg := opts.BeaconConfig
	beaconCfg.FirstExecutionBlockIndex = latestBlockHeader.Number.Uint64()
	beaconCfg.ChainID = chainID.Uint64()
	beaconCfg.GenesisTime = time.Unix(int64(latestBlockHeader.Time), 0)

	// Make the Beacon client manager
	beaconMockServer, err := bnserver.NewBeaconMockServer(logger, *opts.BeaconHostname, *opts.BeaconPort, beaconCfg)
	if err != nil {
		err2 := fsManager.Close()
		if err2 != nil {
			logger.Error("error closing FS manager", "err", err)
		}
		return nil, fmt.Errorf("error creating Beacon mock server: %w", err)
	}
	bnWg := &sync.WaitGroup{}
	beaconMockServer.Start(bnWg)
	beaconNode := client.NewStandardClient(
		client.NewBeaconHttpProvider(
			fmt.Sprintf("http://%s:%d", *opts.BeaconHostname, *opts.BeaconPort), time.Second*30,
		),
	)

	// Make a Docker client mock
	docker := docker.NewDockerMockManager(logger)

	m := &TestManager{
		logger:             logger,
		hardhatRpcClient:   hardhatRpcClient,
		executionClient:    primaryEc,
		beaconMockServer:   beaconMockServer,
		beaconNode:         beaconNode,
		bnWg:               bnWg,
		docker:             docker,
		chainID:            beaconCfg.ChainID,
		fsManager:          fsManager,
		snapshotServiceMap: map[string]Service{},
	}

	// Create the baseline snapshot
	baselineSnapshotID, err := m.takeSnapshot(Service_All)
	if err != nil {
		return nil, fmt.Errorf("error creating baseline snapshot: %w", err)
	}
	m.baselineSnapshotID = baselineSnapshotID

	// Return
	return m, nil
}

// Cleans up the test environment, including the testing folder that houses any generated files
func (m *TestManager) Close() error {
	// Revert to the baseline snapshot
	err := m.revertToSnapshot(m.baselineSnapshotID)
	if err != nil {
		return fmt.Errorf("error reverting to baseline snapshot: %w", err)
	}

	// Shut down the BN
	logger := m.GetLogger()
	if m.bnWg != nil {
		err = m.beaconMockServer.Stop()
		if err != nil {
			logger.Warn("WARNING: nodeset server mock didn't shutdown cleanly", log.Err(err))
		}
		m.bnWg.Wait()
		logger.Info("Stopped Beacon Node mock server")
		m.bnWg = nil
	}

	// Close the filesystem manager
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

func (m *TestManager) GetBeaconMockManager() *bnmanager.BeaconMockManager {
	return m.beaconMockServer.GetManager()
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
	err := m.revertToSnapshot(m.baselineSnapshotID)
	if err != nil {
		return fmt.Errorf("error reverting to baseline snapshot: %w", err)
	}

	// Regenerate the baseline snapshot since Hardhat can't revert to it multiple times
	baselineSnapshotID, err := m.takeSnapshot(Service_All)
	if err != nil {
		return fmt.Errorf("error creating baseline snapshot: %w", err)
	}
	m.baselineSnapshotID = baselineSnapshotID
	return nil
}

// Takes a snapshot of the service states
func (m *TestManager) CreateCustomSnapshot(services Service) (string, error) {
	return m.takeSnapshot(services)
}

// Revert the services to a snapshot state
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
	mgr := m.beaconMockServer.GetManager()
	secondsPerSlot := uint(mgr.GetConfig().SecondsPerSlot)
	err = m.hardhat_increaseTime(secondsPerSlot)
	if err != nil {
		return err
	}

	// Commit the block in the BN
	mgr.CommitBlock(true)
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
	mgr := m.beaconMockServer.GetManager()
	for i := uint(0); i < slots; i++ {
		mgr.CommitBlock(false)
	}

	// Advance the time in Hardhat
	secondsPerSlot := uint(mgr.GetConfig().SecondsPerSlot)
	err := m.hardhatRpcClient.Call(nil, "evm_increaseTime", secondsPerSlot*slots)
	if err != nil {
		return fmt.Errorf("error advancing time on EL: %w", err)
	}
	return nil
}

// Set the highest slot (the head slot) of the Beacon chain, while keeping the local chain head on the client the same.
// Useful for simulating an unsynced client.
func (m *TestManager) SetBeaconHeadSlot(slot uint64) {
	m.beaconMockServer.GetManager().SetHighestSlot(slot)
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

// Takes a snapshot of the service states
func (m *TestManager) takeSnapshot(services Service) (string, error) {
	var snapshotName string
	if services.Contains(Service_EthClients) {
		// Snapshot the EC
		err := m.hardhatRpcClient.Call(&snapshotName, "evm_snapshot")
		if err != nil {
			return "", fmt.Errorf("error creating snapshot: %w", err)
		}

		// Snapshot the BN
		m.beaconMockServer.GetManager().TakeSnapshot(snapshotName)
	}

	// Normally the snapshot name comes from Hardhat but if the EC wasn't snapshotted, make a random one
	if snapshotName == "" {
		for {
			candidate := uuid.New().String()
			_, exists := m.snapshotServiceMap[candidate]
			if !exists {
				snapshotName = candidate
				break
			}
		}
	}

	if services.Contains(Service_Docker) {
		// Snapshot Docker
		err := m.docker.TakeSnapshot(snapshotName)
		if err != nil {
			return "", fmt.Errorf("error creating Docker snapshot: %w", err)
		}
	}

	if services.Contains(Service_Filesystem) {
		// Snapshot the filesystem
		err := m.fsManager.TakeSnapshot(snapshotName)
		if err != nil {
			return "", fmt.Errorf("error creating filesystem snapshot: %w", err)
		}
	}

	// Store the services that were captured
	m.snapshotServiceMap[snapshotName] = services
	return snapshotName, nil
}

// Revert the services to a snapshot state
func (m *TestManager) revertToSnapshot(snapshotID string) error {
	services, exists := m.snapshotServiceMap[snapshotID]
	if !exists {
		return fmt.Errorf("snapshot with ID [%s] does not exist", snapshotID)
	}

	if services.Contains(Service_EthClients) {
		// Revert the EC
		err := m.hardhatRpcClient.Call(nil, "evm_revert", snapshotID)
		if err != nil {
			return fmt.Errorf("error reverting Hardhat to snapshot %s: %w", snapshotID, err)
		}

		// Revert the BN
		err = m.beaconMockServer.GetManager().RevertToSnapshot(snapshotID)
		if err != nil {
			return fmt.Errorf("error reverting the BN to snapshot %s: %w", snapshotID, err)
		}
	}

	if services.Contains(Service_Docker) {
		// Revert Docker
		err := m.docker.RevertToSnapshot(snapshotID)
		if err != nil {
			return fmt.Errorf("error reverting Docker to snapshot %s: %w", snapshotID, err)
		}
	}

	if services.Contains(Service_Filesystem) {
		// Revert the filesystem
		err := m.fsManager.RevertToSnapshot(snapshotID)
		if err != nil {
			return fmt.Errorf("error reverting the filesystem to snapshot %s: %w", snapshotID, err)
		}
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
