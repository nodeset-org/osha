package db

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

const (
	DefaultChainID                      uint64 = 0x90de5e7
	DefaultDepositContractAddressString string = "0xde905175eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
)

var (
	// Default config
	DefaultDepositContractAddress common.Address = common.HexToAddress(DefaultDepositContractAddressString)
)

// Basic Beacon Chain configuration
type Config struct {
	// ==============================
	// === Mock-specific settings ===
	// ==============================

	// Basic settings
	ChainID                      uint64
	SecondsPerSlot               uint64
	SlotsPerEpoch                uint64
	EpochsPerSyncCommitteePeriod uint64
	DepositContract              common.Address

	// Genesis info
	GenesisTime           time.Time
	GenesisForkVersion    []byte
	GenesisValidatorsRoot []byte

	// Altair info
	AltairForkVersion []byte
	AltairForkEpoch   uint64

	// Bellatrix info
	BellatrixForkVersion []byte
	BellatrixForkEpoch   uint64

	// Capella info
	CapellaForkVersion []byte
	CapellaForkEpoch   uint64

	// Deneb info
	DenebForkVersion []byte
	DenebForkEpoch   uint64

	// ==============================
	// === Mock-specific settings ===
	// ==============================

	// The index of the first execution layer block to be linked to in a Beacon chain slot
	FirstExecutionBlockIndex uint64
}

// Creates a new default config instance
func NewDefaultConfig() *Config {
	defaultConfig := &Config{
		ChainID:                      DefaultChainID,
		DepositContract:              DefaultDepositContractAddress,
		SecondsPerSlot:               12,
		SlotsPerEpoch:                32,
		EpochsPerSyncCommitteePeriod: 256,
		GenesisTime:                  time.Now(),
		GenesisForkVersion:           []byte{0x00},
		GenesisValidatorsRoot:        []byte{0x00},
		AltairForkVersion:            common.FromHex("0x90de5e700"),
		AltairForkEpoch:              0,
		BellatrixForkVersion:         common.FromHex("0x90de5e701"),
		BellatrixForkEpoch:           0,
		CapellaForkVersion:           common.FromHex("0x90de5e702"),
		CapellaForkEpoch:             0,
		DenebForkVersion:             common.FromHex("0x90de5e703"),
		DenebForkEpoch:               0,
	}
	return defaultConfig
}

// Clones a config into a new instance
func (c *Config) Clone() *Config {
	return &Config{
		ChainID:                      c.ChainID,
		SecondsPerSlot:               c.SecondsPerSlot,
		SlotsPerEpoch:                c.SlotsPerEpoch,
		EpochsPerSyncCommitteePeriod: c.EpochsPerSyncCommitteePeriod,
		DepositContract:              c.DepositContract,
		GenesisTime:                  c.GenesisTime,
		GenesisForkVersion:           c.GenesisForkVersion,
		GenesisValidatorsRoot:        c.GenesisValidatorsRoot,
		AltairForkVersion:            c.AltairForkVersion,
		AltairForkEpoch:              c.AltairForkEpoch,
		BellatrixForkVersion:         c.BellatrixForkVersion,
		BellatrixForkEpoch:           c.BellatrixForkEpoch,
		CapellaForkVersion:           c.CapellaForkVersion,
		CapellaForkEpoch:             c.CapellaForkEpoch,
		DenebForkVersion:             c.DenebForkVersion,
		DenebForkEpoch:               c.DenebForkEpoch,
	}
}
