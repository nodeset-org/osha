package db

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/rocket-pool/node-manager-core/utils"
	"gopkg.in/yaml.v3"
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
	ChainID                      uint64         `json:"chainID" yaml:"chainID"`
	SecondsPerSlot               uint64         `json:"secondsPerSlot" yaml:"secondsPerSlot"`
	SlotsPerEpoch                uint64         `json:"slotsPerEpoch" yaml:"slotsPerEpoch"`
	EpochsPerSyncCommitteePeriod uint64         `json:"epochsPerSyncCommitteePeriod" yaml:"epochsPerSyncCommitteePeriod"`
	DepositContract              common.Address `json:"depositContract" yaml:"depositContract"`

	// Genesis info
	GenesisTime           time.Time       `json:"genesisTime,omitempty" yaml:"genesisTime,omitempty"`
	GenesisForkVersion    utils.ByteArray `json:"genesisForkVersion" yaml:"genesisForkVersion"`
	GenesisValidatorsRoot utils.ByteArray `json:"genesisValidatorsRoot" yaml:"genesisValidatorsRoot"`

	// Altair info
	AltairForkVersion utils.ByteArray `json:"altairForkVersion" yaml:"altairForkVersion"`
	AltairForkEpoch   uint64          `json:"altairForkEpoch" yaml:"altairForkEpoch"`

	// Bellatrix info
	BellatrixForkVersion utils.ByteArray `json:"bellatrixForkVersion" yaml:"bellatrixForkVersion"`
	BellatrixForkEpoch   uint64          `json:"bellatrixForkEpoch" yaml:"bellatrixForkEpoch"`

	// Capella info
	CapellaForkVersion utils.ByteArray `json:"capellaForkVersion" yaml:"capellaForkVersion"`
	CapellaForkEpoch   uint64          `json:"capellaForkEpoch" yaml:"capellaForkEpoch"`

	// Deneb info
	DenebForkVersion utils.ByteArray `json:"denebForkVersion" yaml:"denebForkVersion"`
	DenebForkEpoch   uint64          `json:"denebForkEpoch" yaml:"denebForkEpoch"`

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
		GenesisTime:                  time.Now().Truncate(time.Second),
		GenesisForkVersion:           common.FromHex("0x90de5e70"),
		GenesisValidatorsRoot:        common.FromHex("0x90de5e70615a7f7115e2b6aac319c03529df8242ae705fba9df39b79c59fa8b0"), // Almost the same as Holesky
		AltairForkVersion:            common.FromHex("0x90de5e71"),
		AltairForkEpoch:              0,
		BellatrixForkVersion:         common.FromHex("0x90de5e72"),
		BellatrixForkEpoch:           0,
		CapellaForkVersion:           common.FromHex("0x90de5e73"),
		CapellaForkEpoch:             0,
		DenebForkVersion:             common.FromHex("0x90de5e74"),
		DenebForkEpoch:               0,
	}
	return defaultConfig
}

// Creates a new config instance from a file
func LoadFromFile(path string) (*Config, error) {
	// Make sure the file exists
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("config file [%s] does not exist", path)
	}

	// Read the file
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file [%s]: %w", path, err)
	}

	// Unmarshal the config
	var config Config
	switch filepath.Ext(path) {
	case ".json":
		err = json.Unmarshal(bytes, &config)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(bytes, &config)
	}
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config file [%s]: %w", path, err)
	}

	// Update fields
	if config.GenesisTime.IsZero() {
		config.GenesisTime = time.Now().Truncate(time.Second)
	}

	return &config, nil
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
