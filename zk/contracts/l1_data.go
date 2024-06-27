package contracts

import (
	"github.com/ledgerwatch/erigon/core/types"
)

// UpdateZkEVMVersion represents a UpdateZkEVMVersion event raised by the Polygonzkevm contract.
type UpdateZkEVMVersion struct {
	NumBatch uint64
	ForkID   uint64
	Version  string
	Raw      types.Log // Blockchain specific contextual infos
}

// SequencedBatchPreEtrog represents a SequenceBatches event raised by the Oldpolygonzkevm contract.
type SequencedBatchPreEtrog struct {
	NumBatch uint64
	Raw      types.Log // Blockchain specific contextual infos
}
