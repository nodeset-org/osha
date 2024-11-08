package internal

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/nodeset-org/osha/vc/db"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/utils"
)

const (
	Pubkey0String string = "0x90de5e70beac090000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	Pubkey1String string = "0x90de5e70beac090000000000000000000000000000000000000000000000000000000000000000000000000000000001"
	Pubkey2String string = "0x90de5e70beac090000000000000000000000000000000000000000000000000000000000000000000000000000000002"

	// Alternate values for testing
	AltFeeRecipientString string = "0x90de5e70fee50000000000000000000000000001"
	AltGraffiti           string = "OSHA alt"
)

var (
	Pubkey0, _      = beacon.HexToValidatorPubkey(Pubkey0String)
	Pubkey1, _      = beacon.HexToValidatorPubkey(Pubkey1String)
	Pubkey2, _      = beacon.HexToValidatorPubkey(Pubkey2String)
	AltFeeRecipient = common.HexToAddress(AltFeeRecipientString)
)

// Provision the standard database for testing
func ProvisionDatabase(kmDB *db.KeyManagerDatabase) {
	// Keystores
	keystores := []*beacon.ValidatorKeystore{
		{
			Crypto:  map[string]any{},
			Name:    "validator0",
			Pubkey:  Pubkey0,
			Version: 4,
			UUID:    uuid.MustParse("00000000-0000-0000-0000-000000000000"),
			Path:    "m/12381/3600/0/0/0",
		}, {
			Crypto:  map[string]any{},
			Name:    "validator1",
			Pubkey:  Pubkey1,
			Version: 4,
			UUID:    uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			Path:    "m/12381/3600/1/0/0",
		}, {
			Crypto:  map[string]any{},
			Name:    "validator2",
			Pubkey:  Pubkey2,
			Version: 4,
			UUID:    uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			Path:    "m/12381/3600/2/0/0",
		},
	}
	passwords := []string{
		"password0",
		"password1",
		"password2",
	}
	slashingData := beacon.SlashingProtectionData{
		Metadata: struct {
			InterchangeFormatVersion utils.Uinteger "json:\"interchange_format_version\""
			GenesisValidatorsRoot    common.Hash    "json:\"genesis_validators_root\""
		}{
			InterchangeFormatVersion: 5,
			GenesisValidatorsRoot:    common.HexToHash(kmDB.GetGenesisValidatorsRoot().Hex()),
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
				Pubkey: Pubkey1,
				SignedBlocks: []struct {
					Slot        utils.Uinteger "json:\"slot\""
					SigningRoot common.Hash    "json:\"signing_root,omitempty\""
				}{
					struct {
						Slot        utils.Uinteger "json:\"slot\""
						SigningRoot common.Hash    "json:\"signing_root,omitempty\""
					}{
						Slot:        0,
						SigningRoot: common.HexToHash("0x51691961200700"),
					},
					struct {
						Slot        utils.Uinteger "json:\"slot\""
						SigningRoot common.Hash    "json:\"signing_root,omitempty\""
					}{
						Slot:        1,
						SigningRoot: common.HexToHash("0x51691961200701"),
					},
				},
				SignedAttestations: []struct {
					SourceEpoch utils.Uinteger "json:\"source_epoch\""
					TargetEpoch utils.Uinteger "json:\"target_epoch\""
					SigningRoot common.Hash    "json:\"signing_root,omitempty\""
				}{
					struct {
						SourceEpoch utils.Uinteger "json:\"source_epoch\""
						TargetEpoch utils.Uinteger "json:\"target_epoch\""
						SigningRoot common.Hash    "json:\"signing_root,omitempty\""
					}{
						SourceEpoch: utils.Uinteger(0),
						TargetEpoch: utils.Uinteger(1),
						SigningRoot: common.HexToHash("0x51691961200702"),
					},
					struct {
						SourceEpoch utils.Uinteger "json:\"source_epoch\""
						TargetEpoch utils.Uinteger "json:\"target_epoch\""
						SigningRoot common.Hash    "json:\"signing_root,omitempty\""
					}{
						SourceEpoch: utils.Uinteger(1),
						TargetEpoch: utils.Uinteger(2),
						SigningRoot: common.HexToHash("0x51691961200703"),
					},
				},
			},
		},
	}

	kmDB.AddValidators(keystores, passwords, &slashingData)
	kmDB.SetFeeRecipient(Pubkey1, AltFeeRecipient)
	kmDB.SetGraffiti(Pubkey2, AltGraffiti)
}
