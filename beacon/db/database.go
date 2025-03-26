package db

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/beacon"
)

// Beacon mock database
type Database struct {
	// Validators registered with the network
	validators []*Validator

	// Lookup of validators by pubkey
	validatorPubkeyMap map[string]*Validator

	// Map of slot indices to execution block indices
	executionBlockMap map[uint64]uint64

	slots            map[uint64]*Slot
	slotBlockRootMap map[common.Hash]*Slot

	// Current slot
	currentSlot uint64

	// Highest slot
	highestSlot uint64

	// Internal fields
	logger                  *slog.Logger
	lock                    *sync.Mutex
	nextExecutionBlockIndex uint64
}

// Create a new database instance
func NewDatabase(logger *slog.Logger, firstExecutionBlockIndex uint64) *Database {
	return &Database{
		logger:                  logger,
		lock:                    &sync.Mutex{},
		nextExecutionBlockIndex: firstExecutionBlockIndex,
		validators:              []*Validator{},
		validatorPubkeyMap:      make(map[string]*Validator),
		executionBlockMap:       make(map[uint64]uint64),
		slots:                   make(map[uint64]*Slot),
		slotBlockRootMap:        make(map[common.Hash]*Slot),
	}
}

// Add a new validator to the database. Returns an error if the validator already exists.
func (db *Database) AddValidator(pubkey beacon.ValidatorPubkey, withdrawalCredentials common.Hash) (*Validator, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	if _, exists := db.validatorPubkeyMap[pubkey.Hex()]; exists {
		return nil, fmt.Errorf("validator with pubkey %s already exists", pubkey.HexWithPrefix())
	}

	index := len(db.validators)
	validator := NewValidator(pubkey, withdrawalCredentials, uint64(index))
	db.validators = append(db.validators, validator)
	db.validatorPubkeyMap[pubkey.Hex()] = validator
	return validator, nil
}

// Add a new slot block header to the database or updates an existing slot if it already exists
func (db *Database) SetSlotBlockRoot(slot uint64, root common.Hash) (bool, error) {

	foundSlot := db.GetSlotByIndex(slot)

	db.lock.Lock()
	defer db.lock.Unlock()

	if foundSlot == nil {
		newSlot := NewSlot(slot, root, 0)
		db.slots[slot] = newSlot
		db.slotBlockRootMap[root] = newSlot
	} else {
		foundSlot.BlockRoot = root
		delete(db.slotBlockRootMap, foundSlot.BlockRoot)
		db.slotBlockRootMap[root] = foundSlot
	}

	return true, nil
}

// Add a new slot exeuction block number to the database or updates an existing slot if it already exists
func (db *Database) SetSlotExecutionBlockNumber(slot uint64, blockNumber uint64) (bool, error) {

	foundSlot := db.GetSlotByIndex(slot)

	db.lock.Lock()
	defer db.lock.Unlock()

	if foundSlot == nil {
		newSlot := NewSlot(slot, common.HexToHash("0x00"), blockNumber)
		db.slots[slot] = newSlot
	} else {
		foundSlot.ExecutionBlockNumber = blockNumber
	}

	return true, nil
}

// Get a validator by its index. Returns nil if it doesn't exist.
func (db *Database) GetValidatorByIndex(index uint) *Validator {
	db.lock.Lock()
	defer db.lock.Unlock()

	dbLength := len(db.validators)
	if index >= uint(dbLength) {
		return nil
	}

	return db.validators[index]
}

// Get a validator by its pubkey. Returns nil if it doesn't exist.
func (db *Database) GetValidatorByPubkey(pubkey beacon.ValidatorPubkey) *Validator {
	db.lock.Lock()
	defer db.lock.Unlock()

	return db.validatorPubkeyMap[pubkey.Hex()]
}

// Get a slot by its index. Returns nil if it doesn't exist.
func (db *Database) GetSlotByIndex(index uint64) *Slot {
	db.lock.Lock()
	defer db.lock.Unlock()

	slot, exists := db.slots[index]
	if !exists {
		return nil
	}

	return slot
}

// Get a slot by its block root. Returns nil if it doesn't exist.
func (db *Database) GetSlotByBlockRoot(blockRoot common.Hash) *Slot {
	db.lock.Lock()
	defer db.lock.Unlock()

	slot, exists := db.slotBlockRootMap[blockRoot]
	if !exists {
		return nil
	}

	return slot
}

// Get all validators
func (db *Database) GetAllValidators() []*Validator {
	db.lock.Lock()
	defer db.lock.Unlock()

	return db.validators
}

// Get the latest local head slot
func (db *Database) GetCurrentSlot() uint64 {
	db.lock.Lock()
	defer db.lock.Unlock()

	return db.currentSlot
}

// Get the highest slot on the chain (the actual chain head)
func (db *Database) GetHighestSlot() uint64 {
	db.lock.Lock()
	defer db.lock.Unlock()

	return db.highestSlot
}

// Add a new block to the chain.
// Set slotValidated to true to "propose a block" for the current slot, linking it to the next Execution block's index.
// Set it to false to "miss" the slot, so there was not block proposed for it.
func (db *Database) CommitBlock(slotValidated bool) {
	db.lock.Lock()
	defer db.lock.Unlock()

	if slotValidated {
		db.executionBlockMap[db.currentSlot] = db.nextExecutionBlockIndex
		db.nextExecutionBlockIndex++
	}
	db.currentSlot++
	if db.currentSlot > db.highestSlot {
		db.highestSlot = db.currentSlot
	}
}

// Set the highest slot on the chain - useful for simulating syncing conditions
func (db *Database) SetHighestSlot(slot uint64) {
	db.lock.Lock()
	defer db.lock.Unlock()

	if slot > db.highestSlot {
		db.highestSlot = slot
	}
}

// Clone the database into a new instance
func (db *Database) Clone() *Database {
	db.lock.Lock()
	defer db.lock.Unlock()

	clone := NewDatabase(db.logger, db.nextExecutionBlockIndex)
	clone.currentSlot = db.currentSlot
	clone.highestSlot = db.highestSlot

	cloneValidators := make([]*Validator, len(db.validators))
	for i, validator := range db.validators {
		cloneValidator := validator.Clone()
		cloneValidators[i] = cloneValidator
		clone.validatorPubkeyMap[validator.Pubkey.Hex()] = cloneValidator
	}
	clone.validators = cloneValidators

	for slot, block := range db.executionBlockMap {
		clone.executionBlockMap[slot] = block
	}
	return clone
}
