package keys

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/rocket-pool/node-manager-core/node/validator"
	"github.com/tyler-smith/go-bip39"
	types "github.com/wealdtech/go-eth2-types/v2"
	eth2util "github.com/wealdtech/go-eth2-util"
)

// KeyGenerator is a simple utility for generating EOA and BLS private keys from a mnemonic and derivation paths.
// This is meant to be used in testing scenarios where complete flexibility of derivation paths isn't a priority.
type KeyGenerator struct {
	mnemonic          string
	ethDerivationPath string
	blsDerivationPath string

	seed      []byte
	masterKey *hdkeychain.ExtendedKey
	ethKeys   map[uint]*ecdsa.PrivateKey
	blsKeys   map[uint]*types.BLSPrivateKey
}

// Create a new key generator from explicit values
func NewKeyGenerator(mnemonic string, ethDerivationPath string, blsDerivationPath string) (*KeyGenerator, error) {
	// Check the mnemonic
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("invalid mnemonic '%s'", mnemonic)
	}

	// Initialize BLS support
	if err := validator.InitializeBls(); err != nil {
		return nil, fmt.Errorf("error initializing BLS library: %w", err)
	}

	// Generate the seed
	seed := bip39.NewSeed(mnemonic, "")

	// Create the master key
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, fmt.Errorf("error creating wallet master key: %w", err)
	}

	return &KeyGenerator{
		mnemonic:          mnemonic,
		ethDerivationPath: ethDerivationPath,
		blsDerivationPath: blsDerivationPath,
		seed:              seed,
		masterKey:         masterKey,
		ethKeys:           map[uint]*ecdsa.PrivateKey{},
		blsKeys:           map[uint]*types.BLSPrivateKey{},
	}, nil
}

// Create a new key generator with default values
func NewKeyGeneratorWithDefaults() (*KeyGenerator, error) {
	return NewKeyGenerator(DefaultMnemonic, DefaultEthDerivationPath, DefaultBeaconDerivationPath)
}

// Get the EOA private key for the given index
func (g *KeyGenerator) GetEthPrivateKey(index uint) (*ecdsa.PrivateKey, error) {
	key, exists := g.ethKeys[index]
	if exists {
		return key, nil
	}

	// Get the derived key
	derivedKey, _, err := getDerivedKey(g.masterKey, g.ethDerivationPath, index)
	if err != nil {
		return nil, fmt.Errorf("error getting node wallet derived key: %w", err)
	}

	// Get the private key from it
	privateKey, err := derivedKey.ECPrivKey()
	if err != nil {
		return nil, fmt.Errorf("error getting node wallet private key: %w", err)
	}
	privateKeyECDSA := privateKey.ToECDSA()
	g.ethKeys[index] = privateKeyECDSA
	return privateKeyECDSA, nil
}

// Get the BLS private key for the given index
func (g *KeyGenerator) GetBlsPrivateKey(index uint) (*types.BLSPrivateKey, error) {
	key, exists := g.blsKeys[index]
	if exists {
		return key, nil
	}

	// Get private key
	path := fmt.Sprintf(g.blsDerivationPath, index)
	blsKey, err := eth2util.PrivateKeyFromSeedAndPath(g.seed, path)
	if err != nil {
		return nil, fmt.Errorf("error getting validator private key for [%s]: %w", path, err)
	}

	g.blsKeys[index] = blsKey
	return blsKey, nil
}

// ==========================
// === Internal Functions ===
// ==========================

// Get the derived key & derivation path for the account at the index
func getDerivedKey(masterKey *hdkeychain.ExtendedKey, derivationPath string, index uint) (*hdkeychain.ExtendedKey, uint, error) {
	formattedDerivationPath := fmt.Sprintf(derivationPath, index)

	// Parse derivation path
	path, err := accounts.ParseDerivationPath(formattedDerivationPath)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid node key derivation path '%s': %w", formattedDerivationPath, err)
	}

	// Follow derivation path
	key := masterKey
	for i, n := range path {
		key, err = key.Derive(n)
		if err == hdkeychain.ErrInvalidChild {
			// Start over with the next index
			return getDerivedKey(masterKey, derivationPath, index+1)
		} else if err != nil {
			return nil, 0, fmt.Errorf("invalid child key at depth %d: %w", i, err)
		}
	}

	// Return
	return key, index, nil
}
