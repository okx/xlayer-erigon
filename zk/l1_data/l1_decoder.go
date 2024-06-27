package l1_data

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ledgerwatch/erigon/zkevm/log"
	"strings"

	"github.com/gateway-fm/cdk-erigon-lib/common"
	"github.com/ledgerwatch/erigon/accounts/abi"
	"github.com/ledgerwatch/erigon/crypto"
	"github.com/ledgerwatch/erigon/zk/contracts"
	"github.com/ledgerwatch/erigon/zk/da"
)

type RollupBaseEtrogBatchData struct {
	Transactions         []byte
	ForcedGlobalExitRoot [32]byte
	ForcedTimestamp      uint64
	ForcedBlockHashL1    [32]byte
}

type ValidiumBatchData struct {
	TransactionsHash     [32]byte
	ForcedGlobalExitRoot [32]byte
	ForcedTimestamp      uint64
	ForcedBlockHashL1    [32]byte
}

// PreEtrogForkID6BatchData is an auto generated low-level Go binding around an user-defined struct.
type PreEtrogForkID6BatchData struct {
	Transactions       []byte
	GlobalExitRoot     [32]byte
	Timestamp          uint64
	MinForcedTimestamp uint64
}

func BuildSequencesForRollup(data []byte) ([]RollupBaseEtrogBatchData, error) {
	var sequences []RollupBaseEtrogBatchData
	err := json.Unmarshal(data, &sequences)
	return sequences, err
}

func BuildSequencesForValidium(data []byte, daUrl string) ([]RollupBaseEtrogBatchData, error) {
	var sequences []RollupBaseEtrogBatchData
	var validiumSequences []ValidiumBatchData
	err := json.Unmarshal(data, &validiumSequences)

	if err != nil {
		return nil, err
	}

	for _, validiumSequence := range validiumSequences {
		hash := common.BytesToHash(validiumSequence.TransactionsHash[:])
		data, err := da.GetOffChainData(context.Background(), daUrl, hash)
		if err != nil {
			return nil, err
		}

		actualTransactionsHash := crypto.Keccak256Hash(data)
		if actualTransactionsHash != hash {
			return nil, fmt.Errorf("unable to fetch off chain data for hash %s, got %s intead", hash.String(), actualTransactionsHash.String())
		}

		sequences = append(sequences, RollupBaseEtrogBatchData{
			Transactions:         data,
			ForcedGlobalExitRoot: validiumSequence.ForcedGlobalExitRoot,
			ForcedTimestamp:      validiumSequence.ForcedTimestamp,
			ForcedBlockHashL1:    validiumSequence.ForcedBlockHashL1,
		})
	}

	return sequences, nil
}

func BuildPreEtrogForkID6SequencesForRollup(data []byte) ([]PreEtrogForkID6BatchData, error) {
	var sequences []PreEtrogForkID6BatchData
	err := json.Unmarshal(data, &sequences)
	return sequences, err
}

func DecodeL1BatchData(txData []byte, daUrl string) ([][]byte, common.Address, uint64, error) {
	// we need to know which version of the ABI to use here so lets find it
	idAsString := fmt.Sprintf("%x", txData[:4])
	abiMapped, found := contracts.SequenceBatchesMapping[idAsString]
	if !found {
		return nil, common.Address{}, 0, fmt.Errorf("unknown l1 call data")
	}

	smcAbi, err := abi.JSON(strings.NewReader(abiMapped))
	if err != nil {
		return nil, common.Address{}, 0, err
	}

	method, err := smcAbi.MethodById(txData[:4])
	if err != nil {
		return nil, common.Address{}, 0, err
	}

	// Unpack method inputs
	data, err := method.Inputs.Unpack(txData[4:])
	if err != nil {
		return nil, common.Address{}, 0, err
	}

	var coinbase common.Address
	var limitTimstamp uint64

	switch idAsString {
	case contracts.KeySequenceBatchForkID6PreEtrog:
		cb, ok := data[1].(common.Address)
		if !ok {
			return nil, common.Address{}, 0, fmt.Errorf("expected position 1 in the l1 call data to be address")
		}
		coinbase = cb

	case contracts.KeySequenceBatchesForkID7Etrog:
		cb, ok := data[1].(common.Address)
		if !ok {
			return nil, common.Address{}, 0, fmt.Errorf("expected position 1 in the l1 call data to be address")
		}
		coinbase = cb
	case contracts.KeySequenceBatchesForkID8Elderberry:
		cb, ok := data[3].(common.Address)
		if !ok {
			return nil, common.Address{}, 0, fmt.Errorf("expected position 3 in the l1 call data to be address")
		}
		coinbase = cb
		ts, ok := data[1].(uint64)
		if !ok {
			return nil, common.Address{}, 0, fmt.Errorf("expected position 1 in the l1 call data to be the limit timestamp")
		}
		limitTimstamp = ts
	case contracts.KeySequenceBatchesForkID8ValidiumElderBerry:
		if daUrl == "" {
			return nil, common.Address{}, 0, fmt.Errorf("data availability url is required for validium")
		}
		cb, ok := data[3].(common.Address)
		if !ok {
			return nil, common.Address{}, 0, fmt.Errorf("expected position 3 in the l1 call data to be address")
		}
		coinbase = cb
		ts, ok := data[1].(uint64)
		if !ok {
			return nil, common.Address{}, 0, fmt.Errorf("expected position 1 in the l1 call data to be the limit timestamp")
		}
		limitTimstamp = ts
	default:
		log.Error(fmt.Sprintf("Unknown l1 call data: %s", idAsString))
		return nil, common.Address{}, 0, fmt.Errorf("unknown l1 call data")
	}

	var etrogSequences []RollupBaseEtrogBatchData
	var preEtrogSequences []PreEtrogForkID6BatchData

	bytedata, err := json.Marshal(data[0])
	if err != nil {
		return nil, coinbase, 0, err
	}

	switch idAsString {
	case contracts.KeySequenceBatchForkID6PreEtrog:
		preEtrogSequences, err = BuildPreEtrogForkID6SequencesForRollup(bytedata)
	case contracts.KeySequenceBatchesForkID7Etrog, contracts.KeySequenceBatchesForkID8Elderberry:
		etrogSequences, err = BuildSequencesForRollup(bytedata)
	case contracts.KeySequenceBatchesForkID8ValidiumElderBerry:
		etrogSequences, err = BuildSequencesForValidium(bytedata, daUrl)
	default:
		log.Error(fmt.Sprintf("Unknown l1 call data: %s", idAsString))
	}

	if err != nil {
		return nil, coinbase, 0, err
	}

	if len(preEtrogSequences) > 0 {
		batchL2Datas := make([][]byte, len(preEtrogSequences))
		for idx, sequence := range preEtrogSequences {
			batchL2Datas[idx] = sequence.Transactions
		}
		return batchL2Datas, coinbase, limitTimstamp, err
	}

	if len(etrogSequences) > 0 {
		batchL2Datas := make([][]byte, len(etrogSequences))
		for idx, sequence := range etrogSequences {
			batchL2Datas[idx] = sequence.Transactions
		}
		return batchL2Datas, coinbase, limitTimstamp, err
	}

	return nil, coinbase, 0, fmt.Errorf("no sequences found")
}
