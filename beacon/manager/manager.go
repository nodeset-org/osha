package manager

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/osha/beacon/db"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/beacon/client"
)

// Beacon mock manager
type BeaconMockManager struct {
	client.IBeaconApiProvider

	name     string
	database *db.Database
	config   *db.Config

	// Internal fields
	snapshots map[string]*db.Database
	logger    *slog.Logger
}

// Create a new beacon mock manager instance
func NewBeaconMockManager(logger *slog.Logger, config *db.Config) *BeaconMockManager {
	return &BeaconMockManager{
		database:  db.NewDatabase(logger, config.FirstExecutionBlockIndex),
		config:    config,
		snapshots: map[string]*db.Database{},
		logger:    logger,
	}
}

func (m *BeaconMockManager) GetName() string {
	return m.name
}

func (m *BeaconMockManager) GetRequirements() {
}

// Set the database for the manager directly if you need to custom provision it
func (m *BeaconMockManager) SetDatabase(db *db.Database) {
	m.database = db
}

func (m *BeaconMockManager) TakeSnapshot() (string, error) {
	snapshot := m.database.Clone()
	if snapshot == nil {
		return "", fmt.Errorf("failed to clone database for snapshot")
	}
	timestamp := time.Now().Format("20060102_150405")
	snapshotName := fmt.Sprintf("%s_%s", m.name, timestamp)
	m.snapshots[snapshotName] = snapshot

	return snapshotName, nil
}

func (m *BeaconMockManager) Close() error {
	if m.database != nil {
		err := m.database.Close()
		if err != nil {
			m.logger.Error("Failed to close database", "error", err)
			return fmt.Errorf("failed to close database: %w", err)
		}
	}

	if m.snapshots != nil {
		for snapshotName := range m.snapshots {
			delete(m.snapshots, snapshotName)
		}
	}

	m.database = nil
	m.snapshots = nil
	return nil
}

func (m *BeaconMockManager) RevertToSnapshot(name string) error {
	snapshot, exists := m.snapshots[name]
	if !exists {
		return fmt.Errorf("snapshot %s does not exist", name)
	}
	m.database = snapshot
	m.logger.Info("Reverted to DB snapshot", "name", name)
	return nil
}

// Returns the manager's Beacon config
func (m *BeaconMockManager) GetConfig() *db.Config {
	return m.config
}

// Increments the Beacon chain slot, committing a new "block" to the chain
// Set slotValidated to true to "propose a block" for the current slot, linking it to the next Execution block's index.
// Set it to false to "miss" the slot, so there was not block proposed for it.
func (m *BeaconMockManager) CommitBlock(slotValidated bool) {
	m.database.CommitBlock(slotValidated)
}

// Returns the current Beacon chain slot
func (m *BeaconMockManager) GetCurrentSlot() uint64 {
	return m.database.GetCurrentSlot()
}

// Returns the highest Beacon chain slot (top of the chain head)
func (m *BeaconMockManager) GetHighestSlot() uint64 {
	return m.database.GetHighestSlot()
}

// Sets the highest slot on the chain - useful for simulating syncing conditions
func (m *BeaconMockManager) SetHighestSlot(slot uint64) {
	m.database.SetHighestSlot(slot)
}

// Add a validator to the Beacon chain
func (m *BeaconMockManager) AddValidator(pubkey beacon.ValidatorPubkey, withdrawalCredentials common.Hash) (*db.Validator, error) {
	return m.database.AddValidator(pubkey, withdrawalCredentials)
}

// Gets a validator by its index or pubkey
func (m *BeaconMockManager) GetValidator(id string) (*db.Validator, error) {
	if len(id) == beacon.ValidatorPubkeyLength*2 || strings.HasPrefix(id, "0x") {
		pubkey, err := beacon.HexToValidatorPubkey(id)
		if err != nil {
			return nil, fmt.Errorf("error parsing pubkey [%s]: %v", id, err)
		}
		return m.database.GetValidatorByPubkey(pubkey), nil
	}
	index, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("error parsing index [%s]: %v", id, err)
	}
	return m.database.GetValidatorByIndex(uint(index)), nil
}

// Gets multiple validators by their indices or pubkeys
func (m *BeaconMockManager) GetValidators(ids []string) ([]*db.Validator, error) {
	if len(ids) == 0 {
		return m.database.GetAllValidators(), nil
	}

	validators := []*db.Validator{}
	for _, id := range ids {
		validator, err := m.GetValidator(id)
		if err != nil {
			return nil, err
		}
		if validator == nil {
			continue
		}
		validators = append(validators, validator)
	}
	return validators, nil
}
