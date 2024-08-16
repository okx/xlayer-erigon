package l1_data

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"encoding/binary"
	"github.com/gateway-fm/cdk-erigon-lib/common"
	"github.com/gateway-fm/cdk-erigon-lib/common/length"
	"github.com/ledgerwatch/erigon/accounts/abi"
	"github.com/ledgerwatch/erigon/crypto"
	"github.com/ledgerwatch/erigon/zk/contracts"
	"github.com/ledgerwatch/erigon/zk/da"
	"github.com/ledgerwatch/erigon/zk/hermez_db"
	zktx "github.com/ledgerwatch/erigon/zk/tx"
	"github.com/ledgerwatch/log/v3"
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
			log.Error(fmt.Sprintf("GetOffChainData, hash:%v", hash.String()))
			return nil, err
		}

		actualTransactionsHash := crypto.Keccak256Hash(data)
		//log.Info(fmt.Sprintf("BuildSequencesForValidium, data len:%v, hash:%v", len(data), hash.String()))
		if actualTransactionsHash != hash {
			log.Info(fmt.Sprintf("BuildSequencesForValidium-----7"))
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

	isValidium := false
	switch idAsString {
	case contracts.SequenceBatchesIdv5_0:
		cb, ok := data[1].(common.Address)
		if !ok {
			return nil, common.Address{}, 0, fmt.Errorf("expected position 1 in the l1 call data to be address")
		}
		coinbase = cb
	case contracts.SequenceBatchesIdv6_6:
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
	case contracts.SequenceBatchesValidiumElderBerry:
		if daUrl == "" {
			log.Info(fmt.Sprintf("SequenceBatchesValidiumElderBerry-----2"))
			return nil, common.Address{}, 0, fmt.Errorf("data availability url is required for validium")
		}
		isValidium = true
		cb, ok := data[3].(common.Address)
		if !ok {
			log.Info(fmt.Sprintf("SequenceBatchesValidiumElderBerry-----4"))
			return nil, common.Address{}, 0, fmt.Errorf("expected position 3 in the l1 call data to be address")
		}
		coinbase = cb
		ts, ok := data[1].(uint64)
		if !ok {
			log.Info(fmt.Sprintf("SequenceBatchesValidiumElderBerry-----6"))
			return nil, common.Address{}, 0, fmt.Errorf("expected position 1 in the l1 call data to be the limit timestamp")
		}
		limitTimstamp = ts
	default:
		return nil, common.Address{}, 0, fmt.Errorf("unknown l1 call data")
	}
	var sequences []RollupBaseEtrogBatchData

	bytedata, err := json.Marshal(data[0])
	if err != nil {
		return nil, coinbase, 0, err
	}

	if isValidium {
		sequences, err = BuildSequencesForValidium(bytedata, daUrl)
	} else {
		sequences, err = BuildSequencesForRollup(bytedata)
	}

	if err != nil {
		log.Info(fmt.Sprintf("SequenceBatchesValidiumElderBerry-----11"))
		return nil, coinbase, 0, err
	}

	batchL2Datas := make([][]byte, len(sequences))
	for idx, sequence := range sequences {
		batchL2Datas[idx] = sequence.Transactions
	}
	//log.Info(fmt.Sprintf("SequenceBatchesValidiumElderBerry-----13, batchL2Datas len:%v", len(sequences)))
	return batchL2Datas, coinbase, limitTimstamp, err
}

type DecodedL1Data struct {
	DecodedData     []zktx.DecodedBatchL2Data
	Coinbase        common.Address
	L1InfoRoot      common.Hash
	IsWorkRemaining bool
	LimitTimestamp  uint64
}

func BreakDownL1DataByBatch(batchNo uint64, forkId uint64, reader *hermez_db.HermezDbReader) (DecodedL1Data, error) {
	decoded := DecodedL1Data{}
	// we expect that the batch we're going to load in next should be in the db already because of the l1 block sync
	// stage, if it is not there we need to panic as we're in a bad state
	batchData, err := reader.GetL1BatchData(batchNo)
	if err != nil {
		log.Error("Error getting batch data", "batch", batchNo, "error", err)
		return decoded, err
	}
	if len(batchData) == 0 {
		log.Info(fmt.Sprintf("BreakDownL1DataByBatch is 0, form GetL1BatchData:%v", batchNo))
		// end of the line for batch recovery so return empty
		return decoded, nil
	}
	decoded.Coinbase = common.BytesToAddress(batchData[:length.Addr])
	decoded.L1InfoRoot = common.BytesToHash(batchData[length.Addr : length.Addr+length.Hash])
	tsBytes := batchData[length.Addr+length.Hash : length.Addr+length.Hash+8]
	decoded.LimitTimestamp = binary.BigEndian.Uint64(tsBytes)
	batchData = batchData[length.Addr+length.Hash+8:]
	decoded.DecodedData, err = zktx.DecodeBatchL2Blocks(batchData, forkId)
	if err != nil {
		log.Error("Error decoding batch data", "batch", batchNo, "error", err)
		return decoded, err
	}
	// no data means no more work to do - end of the line
	if len(decoded.DecodedData) == 0 {
		log.Info("No data in batch", "batch", batchNo)
		return decoded, nil
	}

	decoded.IsWorkRemaining = true
	transactionsInBatch := 0
	for _, batch := range decoded.DecodedData {
		transactionsInBatch += len(batch.Transactions)
	}
	if transactionsInBatch == 0 {
		// we need to check if this batch should simply be empty or not so we need to check against the
		// highest known batch number to see if we have work to do still
		highestKnown, err := reader.GetLastL1BatchData()
		if err != nil {
			return decoded, err
		}
		if batchNo >= highestKnown {
			decoded.IsWorkRemaining = false
		}
	}

	return decoded, err
}
