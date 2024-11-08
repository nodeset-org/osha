package db

import (
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	beacondb "github.com/nodeset-org/osha/beacon/db"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/node/validator/keymanager"
	"github.com/rocket-pool/node-manager-core/utils"
)

const (
	DefaultFeeRecipientString string = "0x90de5e70fee50000000000000000000000000000"
	DefaultGraffiti           string = "OSHA"
	DefaultJwtSecret          string = "osha-default-jwt-secret-01234567"
)

var (
	DefaultFeeRecipient common.Address = common.HexToAddress(DefaultFeeRecipientString)
)

// Info about a validator stored in the key manager
type ValidatorInfo struct {
	Pubkey             beacon.ValidatorPubkey
	DerivationPath     string
	FeeRecipient       common.Address
	Graffiti           string
	SlashingProtection *beacon.SlashingProtectionData
}

// Options for the key manager database
type KeyManagerDatabaseOptions struct {
	DefaultFeeRecipient   *common.Address
	DefaultGraffiti       *string
	GenesisValidatorsRoot *common.Hash
	JwtSecret             *string
}

// Underlying database for the key manager
type KeyManagerDatabase struct {
	logger *slog.Logger
	keys   map[beacon.ValidatorPubkey]ValidatorInfo

	defaultFeeRecipient   common.Address
	defaultGraffiti       string
	genesisValidatorsRoot common.Hash
	jwtSecret             string
}

// Create a new key manager database
func NewKeyManagerDatabase(logger *slog.Logger, opts KeyManagerDatabaseOptions) *KeyManagerDatabase {
	if opts.DefaultFeeRecipient == nil {
		opts.DefaultFeeRecipient = &DefaultFeeRecipient
	}
	if opts.DefaultGraffiti == nil {
		defaultGraffiti := DefaultGraffiti
		opts.DefaultGraffiti = &defaultGraffiti
	}
	if opts.GenesisValidatorsRoot == nil {
		opts.GenesisValidatorsRoot = &beacondb.DefaultGenesisValidatorsRoot
	}
	if opts.JwtSecret == nil {
		defaultJwtSecret := DefaultJwtSecret
		opts.JwtSecret = &defaultJwtSecret
	}
	return &KeyManagerDatabase{
		logger:                logger,
		keys:                  make(map[beacon.ValidatorPubkey]ValidatorInfo),
		defaultFeeRecipient:   *opts.DefaultFeeRecipient,
		defaultGraffiti:       *opts.DefaultGraffiti,
		genesisValidatorsRoot: *opts.GenesisValidatorsRoot,
		jwtSecret:             *opts.JwtSecret,
	}
}

// Create a copy of the database
func (db *KeyManagerDatabase) Clone() *KeyManagerDatabase {
	clone := NewKeyManagerDatabase(db.logger, KeyManagerDatabaseOptions{})
	clone.defaultFeeRecipient = db.defaultFeeRecipient
	clone.defaultGraffiti = db.defaultGraffiti
	clone.genesisValidatorsRoot = db.genesisValidatorsRoot
	clone.jwtSecret = db.jwtSecret

	// Copy the validators
	for pubkey, info := range db.keys {
		cloneInfo := ValidatorInfo{
			Pubkey:         info.Pubkey,
			DerivationPath: info.DerivationPath,
			FeeRecipient:   info.FeeRecipient,
			Graffiti:       info.Graffiti,
		}
		if info.SlashingProtection != nil {
			cloneInfo.SlashingProtection = &beacon.SlashingProtectionData{
				Metadata: info.SlashingProtection.Metadata,
			}
			cloneInfo.SlashingProtection.Data = append(cloneInfo.SlashingProtection.Data, info.SlashingProtection.Data...)
		}
		clone.keys[pubkey] = cloneInfo
	}
	return clone
}

// Get the default fee recipient for new validators
func (db *KeyManagerDatabase) GetDefaultFeeRecipient() common.Address {
	return db.defaultFeeRecipient
}

// Set the default fee recipient for new validators
func (db *KeyManagerDatabase) SetDefaultFeeRecipient(address common.Address) {
	db.defaultFeeRecipient = address
}

// Get the default graffiti for new validators
func (db *KeyManagerDatabase) GetDefaultGraffiti() string {
	return db.defaultGraffiti
}

// Set the default graffiti for new validators
func (db *KeyManagerDatabase) SetDefaultGraffiti(graffiti string) {
	db.defaultGraffiti = graffiti
}

// Get the genesis validators root
func (db *KeyManagerDatabase) GetGenesisValidatorsRoot() common.Hash {
	return db.genesisValidatorsRoot
}

// Set the genesis validators root
func (db *KeyManagerDatabase) SetGenesisValidatorsRoot(genesisValidatorsRoot common.Hash) {
	db.genesisValidatorsRoot = genesisValidatorsRoot
}

// Get the JWT secret
func (db *KeyManagerDatabase) GetJwtSecret() string {
	return db.jwtSecret
}

// Set the JWT secret
func (db *KeyManagerDatabase) SetJwtSecret(jwtSecret string) {
	db.jwtSecret = jwtSecret
}

// Add validators to the database
func (db *KeyManagerDatabase) AddValidators(keystores []*beacon.ValidatorKeystore, passwords []string, slashingProtection *beacon.SlashingProtectionData) []keymanager.ImportKeystoreData {
	results := make([]keymanager.ImportKeystoreData, len(keystores))

	for i, ks := range keystores {
		pubkey := ks.Pubkey

		// Handle duplicate keys
		_, exists := db.keys[pubkey]
		if exists {
			results[i] = keymanager.ImportKeystoreData{
				Status: keymanager.ImportKeystoreStatus_Duplicate,
			}
			continue
		}

		// Make a new info
		info := ValidatorInfo{
			Pubkey:         pubkey,
			DerivationPath: ks.Path,
			FeeRecipient:   db.defaultFeeRecipient,
			Graffiti:       db.defaultGraffiti,
		}

		// Make a new slashing protection if it's in there
		found := false
		if slashingProtection != nil {
			for _, data := range slashingProtection.Data {
				if data.Pubkey != pubkey {
					continue
				}
				found = true
				info.SlashingProtection = &beacon.SlashingProtectionData{
					Metadata: slashingProtection.Metadata,
				}
				info.SlashingProtection.Data = append(info.SlashingProtection.Data, data)
			}
		}
		if !found {
			info.SlashingProtection = &beacon.SlashingProtectionData{
				Metadata: struct {
					InterchangeFormatVersion utils.Uinteger "json:\"interchange_format_version\""
					GenesisValidatorsRoot    common.Hash    "json:\"genesis_validators_root\""
				}{
					InterchangeFormatVersion: 5,
					GenesisValidatorsRoot:    db.genesisValidatorsRoot,
				},
				Data: []struct {
					Pubkey       beacon.ValidatorPubkey "json:\"pubkey\""
					SignedBlocks []struct {
						Slot        utils.Uinteger "json:\"slot\""
						SigningRoot common.Hash    "json:\"signing_root,omitempty\""
					} "json:\"signed_blocks\""
					SignedAttestations []struct {
						SourceEpoch utils.Uinteger "json:\"source_epoch\""
						TargetEpoch utils.Uinteger "json:\"target_epoch\""
						SigningRoot common.Hash    "json:\"signing_root,omitempty\""
					} "json:\"signed_attestations\""
				}{
					struct {
						Pubkey       beacon.ValidatorPubkey "json:\"pubkey\""
						SignedBlocks []struct {
							Slot        utils.Uinteger "json:\"slot\""
							SigningRoot common.Hash    "json:\"signing_root,omitempty\""
						} "json:\"signed_blocks\""
						SignedAttestations []struct {
							SourceEpoch utils.Uinteger "json:\"source_epoch\""
							TargetEpoch utils.Uinteger "json:\"target_epoch\""
							SigningRoot common.Hash    "json:\"signing_root,omitempty\""
						} "json:\"signed_attestations\""
					}{
						Pubkey: pubkey,
						SignedBlocks: []struct {
							Slot        utils.Uinteger "json:\"slot\""
							SigningRoot common.Hash    "json:\"signing_root,omitempty\""
						}{},
						SignedAttestations: []struct {
							SourceEpoch utils.Uinteger "json:\"source_epoch\""
							TargetEpoch utils.Uinteger "json:\"target_epoch\""
							SigningRoot common.Hash    "json:\"signing_root,omitempty\""
						}{},
					},
				},
			}
		}
		db.keys[pubkey] = info
		results[i] = keymanager.ImportKeystoreData{
			Status: keymanager.ImportKeystoreStatus_Imported,
		}
	}

	return results
}

// Get the info for all validators
func (db *KeyManagerDatabase) GetAllValidators() []keymanager.GetKeystoreData {
	results := make([]keymanager.GetKeystoreData, len(db.keys))
	i := 0
	for _, info := range db.keys {
		results[i] = keymanager.GetKeystoreData{
			Pubkey:         info.Pubkey,
			DerivationPath: info.DerivationPath,
			ReadOnly:       false,
		}
		i++
	}
	return results
}

// Delete validators from the database
func (db *KeyManagerDatabase) DeleteValidators(pubkeys []beacon.ValidatorPubkey) ([]keymanager.DeleteKeystoreData, *beacon.SlashingProtectionData) {
	results := make([]keymanager.DeleteKeystoreData, len(pubkeys))
	totalSlashingProtection := beacon.SlashingProtectionData{
		Metadata: struct {
			InterchangeFormatVersion utils.Uinteger "json:\"interchange_format_version\""
			GenesisValidatorsRoot    common.Hash    "json:\"genesis_validators_root\""
		}{
			InterchangeFormatVersion: 5,
			GenesisValidatorsRoot:    db.genesisValidatorsRoot,
		},
	}

	for i, pubkey := range pubkeys {
		// Check if the key exists
		info, exists := db.keys[pubkey]
		if !exists {
			results[i] = keymanager.DeleteKeystoreData{
				Status: keymanager.DeleteKeystoreStatus_NotFound,
			}
			continue
		}

		// Delete the key
		if info.SlashingProtection != nil {
			totalSlashingProtection.Data = append(totalSlashingProtection.Data, info.SlashingProtection.Data...)
		}
		delete(db.keys, pubkey)
		results[i] = keymanager.DeleteKeystoreData{
			Status: keymanager.DeleteKeystoreStatus_Deleted,
		}
	}
	return results, &totalSlashingProtection
}

// Get the fee recipient for a validator
func (db *KeyManagerDatabase) GetFeeRecipient(pubkey beacon.ValidatorPubkey) keymanager.GetFeeRecipientData {
	info, exists := db.keys[pubkey]
	if !exists {
		return keymanager.GetFeeRecipientData{
			Pubkey:     pubkey,
			EthAddress: db.defaultFeeRecipient,
		}
	}
	return keymanager.GetFeeRecipientData{
		Pubkey:     pubkey,
		EthAddress: info.FeeRecipient,
	}
}

// Set the fee recipient for a validator
func (db *KeyManagerDatabase) SetFeeRecipient(pubkey beacon.ValidatorPubkey, address common.Address) bool {
	info, exists := db.keys[pubkey]
	if !exists {
		return false
	}
	info.FeeRecipient = address
	db.keys[pubkey] = info
	return true
}

// Get the graffiti for a validator
func (db *KeyManagerDatabase) GetGraffiti(pubkey beacon.ValidatorPubkey) keymanager.GetGraffitiData {
	info, exists := db.keys[pubkey]
	if !exists {
		return keymanager.GetGraffitiData{
			Pubkey:   pubkey,
			Graffiti: db.defaultGraffiti,
		}
	}
	return keymanager.GetGraffitiData{
		Pubkey:   pubkey,
		Graffiti: info.Graffiti,
	}
}

// Set the graffiti for a validator
func (db *KeyManagerDatabase) SetGraffiti(pubkey beacon.ValidatorPubkey, graffiti string) bool {
	info, exists := db.keys[pubkey]
	if !exists {
		return false
	}
	info.Graffiti = graffiti
	db.keys[pubkey] = info
	return true
}
