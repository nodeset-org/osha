package db

import (
	"github.com/ethereum/go-ethereum/common"
)

type Slot struct {
	Index                uint64
	BlockRoot            common.Hash
	ExecutionBlockNumber uint64
}

func NewSlot(index uint64, blockRoot common.Hash, executionBlockNumber uint64) *Slot {
	return &Slot{
		Index:                index,
		BlockRoot:            blockRoot,
		ExecutionBlockNumber: executionBlockNumber,
	}
}

func (v *Slot) Clone() *Slot {
	return &Slot{
		BlockRoot:            v.BlockRoot,
		Index:                v.Index,
		ExecutionBlockNumber: v.ExecutionBlockNumber,
	}
}
