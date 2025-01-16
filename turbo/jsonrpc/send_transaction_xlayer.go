package jsonrpc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	libcommon "github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/common/hexutil"
	"github.com/ledgerwatch/erigon-lib/common/hexutility"
	"github.com/ledgerwatch/erigon/accounts/abi"
	"github.com/ledgerwatch/erigon/accounts/abi/bind"
	"github.com/ledgerwatch/erigon/core"
	"github.com/ledgerwatch/erigon/core/state"
	"github.com/ledgerwatch/erigon/core/types"
	"github.com/ledgerwatch/erigon/core/vm"
	"github.com/ledgerwatch/erigon/params"
	"github.com/ledgerwatch/erigon/rpc"
	ethapi2 "github.com/ledgerwatch/erigon/turbo/adapter/ethapi"
	"github.com/ledgerwatch/erigon/turbo/rpchelper"
	"github.com/ledgerwatch/erigon/turbo/transactions"
)

type PackedUserOperation struct {
	Sender             libcommon.Address
	Nonce              *big.Int
	InitCode           []byte
	CallData           []byte
	AccountGasLimits   [32]byte
	PreVerificationGas *big.Int
	GasFees            [32]byte
	PaymasterAndData   []byte
	Signature          []byte
}

var PayABIData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"entryPoint\",\"outputs\":[{\"internalType\":\"contractIEntryPoint\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getNonce\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"initCode\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"callData\",\"type\":\"bytes\"},{\"internalType\":\"bytes32\",\"name\":\"accountGasLimits\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"preVerificationGas\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"gasFees\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"paymasterAndData\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"internalType\":\"structPackedUserOperation\",\"name\":\"userOp\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"userOpHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"missingAccountFunds\",\"type\":\"uint256\"}],\"name\":\"validateUserOp\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"validationData\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

func (api *APIImpl) parseABI(txn types.Transaction, chainId *big.Int) {
	txData := txn.GetData()
	hexData := hexutil.Bytes(txData)
	hexUtilityData := (*hexutility.Bytes)(&hexData)
	if len(txData) < 0 {
		return
	}

	signer := types.LatestSignerForChainID(chainId)
	fromAddress, _ := txn.Sender(*signer)
	log.Info(fmt.Sprintf("hash:%v, sender:%v, to:%v, data:%v", txn.Hash(), fromAddress, txn.GetTo(), hexUtilityData))

	parsedABI, err := abi.JSON(bytes.NewReader([]byte(PayABIData.ABI)))
	if err != nil {
		log.Error("Failed to parse ABI", "error", err)
		return
	}

	method, err := parsedABI.MethodById(txData[:4])
	if err != nil {
		log.Error(fmt.Sprintf("Parse method failed %v", err))
		return
	}

	log.Info(fmt.Sprintf("method: %s", method.Name))

	args := make(map[string]interface{})
	err = method.Inputs.UnpackIntoMap(args, txData[4:])
	if err != nil {
		log.Error("Parse method failed %v", err)
		return
	}

	userOp := args["userOp"].(map[string]interface{})
	parsedUserOp := PackedUserOperation{
		Sender:             userOp["sender"].(libcommon.Address),
		Nonce:              userOp["nonce"].(*big.Int),
		InitCode:           userOp["initCode"].([]byte),
		CallData:           userOp["callData"].([]byte),
		AccountGasLimits:   userOp["accountGasLimits"].([32]byte),
		PreVerificationGas: userOp["preVerificationGas"].(*big.Int),
		GasFees:            userOp["gasFees"].([32]byte),
		PaymasterAndData:   userOp["paymasterAndData"].([]byte),
		Signature:          userOp["signature"].([]byte),
	}
	log.Info("%v", parsedUserOp)

	userOpHash := args["userOpHash"].([32]byte)
	log.Info("UserOpHash: %x\n", userOpHash[:])

	missingAccountFunds := args["missingAccountFunds"].(*big.Int)
	log.Info("MissingAccountFunds: %s\n", missingAccountFunds.String())
}

// EstimateGas implements eth_estimateGas. Returns an estimate of how much gas is necessary to allow the transaction to complete. The transaction will not be added to the blockchain.
func (api *APIImpl) estimateGas(txn types.Transaction, chainId *big.Int) (hexutil.Uint64, error) {
	signer := types.LatestSignerForChainID(chainId)
	fromAddress, err := txn.Sender(*signer)
	fromAddressHex := libcommon.HexToAddress(fromAddress.String())
	gas := txn.GetGas()
	gp := new(big.Int).SetBytes(txn.GetPrice().Bytes())
	newGP := (*hexutil.Big)(gp)
	value := new(big.Int).SetBytes(txn.GetValue().Bytes())
	newValue := (*hexutil.Big)(value)

	nonce := txn.GetNonce()

	data := txn.GetData()
	hexData := hexutil.Bytes(data)
	hexUtilityData := (*hexutility.Bytes)(&hexData)
	argsOrNil := &ethapi2.CallArgs{
		From:     &fromAddressHex,
		To:       txn.GetTo(),
		Gas:      (*hexutil.Uint64)(&gas),
		GasPrice: newGP,
		Value:    newValue,
		Nonce:    (*hexutil.Uint64)(&nonce),
		Data:     hexUtilityData,
		Input:    hexUtilityData,
		ChainID:  (*hexutil.Big)(chainId),
	}

	var args ethapi2.CallArgs
	// if we actually get CallArgs here, we use them
	if argsOrNil != nil {
		args = *argsOrNil
	}

	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	dbtx, err := api.db.BeginRo(ctx)
	if err != nil {
		return 0, err
	}
	defer dbtx.Rollback()

	// Binary search the gas requirement, as it may be higher than the amount used
	var (
		lo     = params.TxGas - 1
		hi     uint64
		gasCap uint64
	)
	// Use zero address if sender unspecified.
	if args.From == nil {
		args.From = new(libcommon.Address)
	}

	bNrOrHash := rpc.BlockNumberOrHashWithNumber(rpc.PendingBlockNumber)

	// Determine the highest gas limit can be used during the estimation.
	if args.Gas != nil && uint64(*args.Gas) >= params.TxGas {
		hi = uint64(*args.Gas)
	} else {
		// Retrieve the block to act as the gas ceiling
		h, err := headerByNumberOrHash(ctx, dbtx, bNrOrHash, api)
		if err != nil {
			return 0, err
		}
		if h == nil {
			// block number not supplied, so we haven't found a pending block, read the latest block instead
			h, err = headerByNumberOrHash(ctx, dbtx, latestNumOrHash, api)
			if err != nil {
				return 0, err
			}
			if h == nil {
				return 0, nil
			}
		}
		hi = h.GasLimit
	}

	var feeCap *big.Int
	if args.GasPrice != nil && (args.MaxFeePerGas != nil || args.MaxPriorityFeePerGas != nil) {
		return 0, errors.New("both gasPrice and (maxFeePerGas or maxPriorityFeePerGas) specified")
	} else if args.GasPrice != nil {
		feeCap = args.GasPrice.ToInt()
	} else if args.MaxFeePerGas != nil {
		feeCap = args.MaxFeePerGas.ToInt()
	} else {
		feeCap = libcommon.Big0
	}
	// Recap the highest gas limit with account's available balance.
	if feeCap.Sign() != 0 {
		cacheView, err := api.stateCache.View(ctx, dbtx)
		if err != nil {
			return 0, err
		}
		stateReader := state.NewCachedReader2(cacheView, dbtx)
		state := state.New(stateReader)
		if state == nil {
			return 0, fmt.Errorf("can't get the current state")
		}

		balance := state.GetBalance(*args.From) // from can't be nil
		available := balance.ToBig()
		if args.Value != nil {
			if args.Value.ToInt().Cmp(available) >= 0 {
				return 0, errors.New("insufficient funds for transfer")
			}
			available.Sub(available, args.Value.ToInt())
		}
		allowance := new(big.Int).Div(available, feeCap)

		// If the allowance is larger than maximum uint64, skip checking
		if allowance.IsUint64() && hi > allowance.Uint64() {
			transfer := args.Value
			if transfer == nil {
				transfer = new(hexutil.Big)
			}
			log.Debug("Gas estimation capped by limited funds", "original", hi, "balance", balance,
				"sent", transfer.ToInt(), "maxFeePerGas", feeCap, "fundable", allowance)
			hi = allowance.Uint64()
		}
	}

	// Recap the highest gas allowance with specified gascap.
	if hi > api.GasCap {
		log.Debug("Caller gas above allowance, capping", "requested", hi, "cap", api.GasCap)
		hi = api.GasCap
	}
	gasCap = hi

	chainConfig, err := api.chainConfig(ctx, dbtx)
	if err != nil {
		return 0, err
	}
	engine := api.engine()

	latestCanBlockNumber, latestCanHash, isLatest, err := rpchelper.GetCanonicalBlockNumber_zkevm(bNrOrHash, dbtx, api.filters) // DoCall cannot be executed on non-canonical blocks
	if err != nil {
		return 0, err
	}

	// try and get the block from the lru cache first then try DB before failing
	block := api.tryBlockFromLru(latestCanHash)
	if block == nil {
		block, err = api.blockWithSenders(ctx, dbtx, latestCanHash, latestCanBlockNumber)
		if err != nil {
			return 0, err
		}
	}
	if block == nil {
		return 0, fmt.Errorf("could not find latest block in cache or db")
	}

	stateReader, err := rpchelper.CreateStateReaderFromBlockNumber(ctx, dbtx, latestCanBlockNumber, isLatest, 0, api.stateCache, api.historyV3(dbtx), chainConfig.ChainName)
	if err != nil {
		return 0, err
	}
	header := block.HeaderNoCopy()

	caller, err := transactions.NewReusableCaller(engine, stateReader, nil, header, args, api.GasCap, latestNumOrHash, dbtx, api._blockReader, chainConfig, api.evmCallTimeout, api.VirtualCountersSmtReduction, false)
	if err != nil {
		return 0, err
	}

	// Create a helper to check if a gas allowance results in an executable transaction
	executable := func(gas uint64) (bool, *core.ExecutionResult, error) {
		result, err := caller.DoCallWithNewGas(ctx, gas)
		if err != nil {
			if errors.Is(err, core.ErrIntrinsicGas) {
				// Special case, raise gas limit
				return true, nil, nil
			}

			// Bail out
			return true, nil, err
		}
		return result.Failed(), result, nil
	}

	// Execute the binary search and hone in on an executable gas limit
	for lo+1 < hi {
		mid := (hi + lo) / 2
		failed, _, err := executable(mid)
		// If the error is not nil(consensus error), it means the provided message
		// call or transaction will never be accepted no matter how much gas it is
		// assigened. Return the error directly, don't struggle any more.
		if err != nil {
			return 0, err
		}
		if failed {
			lo = mid
		} else {
			hi = mid
		}
	}

	// Reject the transaction as invalid if it still fails at the highest allowance
	if hi == gasCap {
		failed, result, err := executable(hi)
		if err != nil {
			return 0, err
		}
		if failed {
			if result != nil && !errors.Is(result.Err, vm.ErrOutOfGas) {
				if len(result.Revert()) > 0 {
					return 0, ethapi2.NewRevertError(result)
				}
				return 0, result.Err
			}
			// Otherwise, the specified gas cap is too low
			return 0, fmt.Errorf("gas required exceeds allowance (%d)", gasCap)
		}
	}
	return hexutil.Uint64(hi), nil
}
