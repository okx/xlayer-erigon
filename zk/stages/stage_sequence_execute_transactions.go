package stages

import (
	"context"
	"os"
	"sync"

	"github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/kv"
	"golang.org/x/sys/cpu"

	"io"

	mapset "github.com/deckarep/golang-set/v2"
	types2 "github.com/ledgerwatch/erigon-lib/types"
	"github.com/ledgerwatch/erigon/core"
	"github.com/ledgerwatch/erigon/core/state"
	"github.com/ledgerwatch/erigon/core/types"
	"github.com/ledgerwatch/erigon/core/vm"
	"github.com/ledgerwatch/erigon/core/vm/evmtypes"
	"github.com/ledgerwatch/erigon/crypto/keccak"
	"github.com/ledgerwatch/erigon/zk/utils"
	"github.com/ledgerwatch/log/v3"
	"github.com/ledgerwatch/secp256k1"
)

func getNextPoolTransactions(ctx context.Context, cfg SequenceBlockCfg, executionAt, forkId uint64, alreadyYielded mapset.Set[[32]byte]) ([]types.Transaction, []common.Hash, bool, error) {
	cfg.txPool.LockFlusher()
	defer cfg.txPool.UnlockFlusher()

	var ids []common.Hash
	var transactions []types.Transaction
	var allConditionsOk bool
	var err error

	gasLimit := utils.GetBlockGasLimitForFork(forkId)

	ti := utils.StartTimer("txpool", "get-transactions")
	defer ti.LogTimer()

	if err := cfg.txPoolDb.View(ctx, func(poolTx kv.Tx) error {
		slots := types2.TxsRlp{}
		if allConditionsOk, _, err = cfg.txPool.YieldBest(cfg.yieldSize, &slots, poolTx, executionAt, gasLimit, 0, alreadyYielded); err != nil {
			return err
		}
		yieldedTxs, yieldedIds, toRemove, err := extractTransactionsFromSlot(&slots, executionAt, cfg)
		if err != nil {
			return err
		}
		for _, txId := range toRemove {
			cfg.txPool.MarkForDiscardFromPendingBest(txId)
		}
		transactions = append(transactions, yieldedTxs...)
		ids = append(ids, yieldedIds...)
		return nil
	}); err != nil {
		return nil, nil, allConditionsOk, err
	}

	return transactions, ids, allConditionsOk, err
}

func getLimboTransaction(ctx context.Context, cfg SequenceBlockCfg, txHash *common.Hash, executionAt uint64) ([]types.Transaction, error) {
	cfg.txPool.LockFlusher()
	defer cfg.txPool.UnlockFlusher()

	var transactions []types.Transaction
	// ensure we don't spin forever looking for transactions, attempt for a while then exit up to the caller
	if err := cfg.txPoolDb.View(ctx, func(poolTx kv.Tx) error {
		slots, err := cfg.txPool.GetLimboTxRplsByHash(poolTx, txHash)
		if err != nil {
			return err
		}

		if slots != nil {
			// ignore the toRemove value here, we know the RLP will be sound as we had to read it from the pool
			// in the first place to get it into limbo
			transactions, _, _, err = extractTransactionsFromSlot(slots, executionAt, cfg)
			if err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return transactions, nil
}

func extractTransactionsFromSlot(slot *types2.TxsRlp, currentHeight uint64, cfg SequenceBlockCfg) ([]types.Transaction, []common.Hash, []common.Hash, error) {
	ids := make([]common.Hash, 0, len(slot.TxIds))
	transactions := make([]types.Transaction, 0, len(slot.Txs))
	toRemove := make([]common.Hash, 0)

	signer := types.MakeSigner(cfg.chainConfig, currentHeight, 0)

	type result struct {
		index       int
		transaction types.Transaction
		id          common.Hash
		toRemove    bool
		err         error
	}

	results := make([]*result, len(slot.Txs))
	var wg sync.WaitGroup

	_, use_avx := os.LookupEnv("ERIGON_USE_AVX")
	if use_avx && (cpu.X86.HasAVX2 || cpu.X86.HasAVX512) {
		chunkSize := 4
		if cpu.X86.HasAVX512 {
			chunkSize = 8
		}
		/*
				log.Info("[extractTransaction] Using AVX512")
			} else {
				log.Info("[extractTransaction] Using AVX2")
			}
		*/
		txs := make([]types.Transaction, len(slot.Txs))
		for idx, txBytes := range slot.Txs {
			res := &result{index: idx}
			transaction, err := types.DecodeTransaction(txBytes)
			if err == io.EOF {
				res.toRemove = true
				results[idx] = res
				continue
			}
			if err != nil {
				res.toRemove = true
				res.id = slot.TxIds[idx]
				res.err = err
				results[idx] = res
				continue
			}
			res.transaction = transaction
			res.id = slot.TxIds[idx]
			results[idx] = res
			txs[idx] = transaction
		}
		wg.Wait()

		n := len(txs)
		signChainID := signer.ChainID().ToBig()
		txHashes := make([]common.Hash, n)

		// first, run chunkSize hashes per core (AVX/AVX512) in parallel
		n1 := (n / chunkSize) * chunkSize
		for idx := 0; idx < n1; idx += chunkSize {
			wg.Add(1)
			go func(idx int, atxs []types.Transaction, hashes []common.Hash) {
				defer wg.Done()
				keccak.SigningHashAVX(signChainID, atxs, hashes, cpu.X86.HasAVX512)
			}(idx, txs[idx:idx+chunkSize], txHashes[idx:idx+chunkSize])
		}
		// run the remaining hashes (<chunkSize) sequentially
		for idx := n1; idx < n; idx += 1 {
			thisChainId := signChainID
			if !txs[idx].Protected() {
				thisChainId = nil
			}
			txHashes[idx] = txs[idx].SigningHash(thisChainId)
		}
		wg.Wait()

		cryptoContext := secp256k1.ContextForThread(1)
		for idx := range txs {
			if !results[idx].toRemove {
				sender, err := signer.SenderWithContext(cryptoContext, txs[idx], &txHashes[idx])
				if err != nil {
					results[idx].toRemove = true
					results[idx].err = err
				} else {
					txs[idx].SetSender(sender)
					results[idx].transaction = txs[idx]
				}
			}
		}
	} else {
		for idx, txBytes := range slot.Txs {
			wg.Add(1)
			go func(idx int, txBytes []byte) {
				defer wg.Done()

				cryptoContext := secp256k1.ContextForThread(1)

				res := &result{index: idx}

				transaction, err := types.DecodeTransaction(txBytes)
				if err == io.EOF {
					res.toRemove = true
					results[idx] = res
					return
				}
				if err != nil {
					res.toRemove = true
					res.id = slot.TxIds[idx]
					res.err = err
					results[idx] = res
					return
				}

				sender, err := signer.SenderWithContext(cryptoContext, transaction, nil)
				if err != nil {
					res.toRemove = true
					res.id = slot.TxIds[idx]
					res.err = err
					results[idx] = res
					return
				}

				transaction.SetSender(sender)
				res.transaction = transaction
				res.id = slot.TxIds[idx]
				results[idx] = res
			}(idx, txBytes)
		}

		wg.Wait()
	}

	for _, res := range results {
		if res.toRemove {
			toRemove = append(toRemove, res.id)
			if res.err != nil {
				log.Warn("[extractTransaction] Failed to process transaction",
					"error", res.err, "id", res.id)
			}
			continue
		}
		transactions = append(transactions, res.transaction)
		ids = append(ids, res.id)
	}

	return transactions, ids, toRemove, nil
}

type overflowType uint8

const (
	overflowNone overflowType = iota
	overflowCounters
	overflowGas
)

func attemptAddTransaction(
	cfg SequenceBlockCfg,
	sdb *stageDb,
	ibs *state.IntraBlockState,
	batchCounters *vm.BatchCounterCollector,
	blockContext *evmtypes.BlockContext,
	header *types.Header,
	transaction types.Transaction,
	effectiveGasPrice uint8,
	l1Recovery bool,
	forkId, l1InfoIndex uint64,
	blockDataSizeChecker *BlockDataChecker,
) (*types.Receipt, *core.ExecutionResult, overflowType, error) {
	var batchDataOverflow, overflow bool
	var err error

	txCounters := vm.NewTransactionCounter(transaction, sdb.smt.GetDepth(), uint16(forkId), cfg.zk.VirtualCountersSmtReduction, cfg.zk.ShouldCountersBeUnlimited(l1Recovery))
	overflow, err = batchCounters.AddNewTransactionCounters(txCounters)

	// run this only once the first time, do not add it on rerun
	if blockDataSizeChecker != nil {
		txL2Data, err := txCounters.GetL2DataCache()
		if err != nil {
			return nil, nil, overflowNone, err
		}
		batchDataOverflow = blockDataSizeChecker.AddTransactionData(txL2Data)
		if batchDataOverflow {
			log.Info("BatchL2Data limit reached. Not adding last transaction", "txHash", transaction.Hash())
		}
	}
	if err != nil {
		return nil, nil, overflowNone, err
	}
	anyOverflow := overflow || batchDataOverflow
	if anyOverflow && !l1Recovery {
		log.Debug("Transaction preexecute overflow detected", "txHash", transaction.Hash(), "counters", batchCounters.CombineCollectorsNoChanges().UsedAsString())
		return nil, nil, overflowCounters, nil
	}

	gasPool := new(core.GasPool).AddGas(transactionGasLimit)

	// set the counter collector on the config so that we can gather info during the execution
	cfg.zkVmConfig.CounterCollector = txCounters.ExecutionCounters()

	// TODO: possibly inject zero tracer here!

	snapshot := ibs.Snapshot()
	ibs.Init(transaction.Hash(), common.Hash{}, 0)

	evm := vm.NewZkEVM(*blockContext, evmtypes.TxContext{}, ibs, cfg.chainConfig, *cfg.zkVmConfig)

	gasUsed := header.GasUsed

	receipt, execResult, _, err := core.ApplyTransaction_zkevm(
		cfg.chainConfig,
		cfg.engine,
		evm,
		gasPool,
		ibs,
		noop,
		header,
		transaction,
		&gasUsed,
		effectiveGasPrice,
		false,
	)

	if err != nil {
		return nil, nil, overflowNone, err
	}

	if err = txCounters.ProcessTx(ibs, execResult.ReturnData); err != nil {
		return nil, nil, overflowNone, err
	}

	batchCounters.UpdateExecutionAndProcessingCountersCache(txCounters)
	// now that we have executed we can check again for an overflow
	if overflow, err = batchCounters.CheckForOverflow(l1InfoIndex != 0); err != nil {
		return nil, nil, overflowNone, err
	}

	counters := batchCounters.CombineCollectorsNoChanges().UsedAsString()
	if overflow {
		log.Debug("Transaction overflow detected", "txHash", transaction.Hash(), "coutners", counters)
		ibs.RevertToSnapshot(snapshot)
		return nil, nil, overflowCounters, nil
	}
	if gasUsed > header.GasLimit {
		log.Debug("Transaction overflows block gas limit", "txHash", transaction.Hash(), "txGas", receipt.GasUsed, "blockGasUsed", header.GasUsed)
		ibs.RevertToSnapshot(snapshot)
		return nil, nil, overflowGas, nil
	}
	log.Debug("Transaction added", "txHash", transaction.Hash(), "coutners", counters)

	// add the gas only if not reverted. This should not be moved above the overflow check
	header.GasUsed = gasUsed

	// we need to keep hold of the effective percentage used
	// todo [zkevm] for now we're hard coding to the max value but we need to calc this properly
	if err = sdb.hermezDb.WriteEffectiveGasPricePercentage(transaction.Hash(), effectiveGasPrice); err != nil {
		return nil, nil, overflowNone, err
	}

	ibs.FinalizeTx(evm.ChainRules(), noop)

	return receipt, execResult, overflowNone, nil
}
