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
	validatorPubkeyMap map[beacon.ValidatorPubkey]*Validator

	// Map of slot indices to execution block indices
	executionBlockMap map[uint64]uint64

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
		validatorPubkeyMap:      make(map[beacon.ValidatorPubkey]*Validator),
		executionBlockMap:       make(map[uint64]uint64),
	}
}

// Add a new validator to the database. Returns an error if the validator already exists.
func (db *Database) AddValidator(pubkey beacon.ValidatorPubkey, withdrawalCredentials common.Hash) (*Validator, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	if _, exists := db.validatorPubkeyMap[pubkey]; exists {
		return nil, fmt.Errorf("validator with pubkey %s already exists", pubkey.HexWithPrefix())
	}

	index := len(db.validators)
	validator := NewValidator(pubkey, withdrawalCredentials, uint64(index))
	db.validators = append(db.validators, validator)
	db.validatorPubkeyMap[pubkey] = validator
	return validator, nil
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

	return db.validatorPubkeyMap[pubkey]
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
		clone.validatorPubkeyMap[validator.Pubkey] = cloneValidator
	}
	clone.validators = cloneValidators

	for slot, block := range db.executionBlockMap {
		clone.executionBlockMap[slot] = block
	}
	return clone
}
