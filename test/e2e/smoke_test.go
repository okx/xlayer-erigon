package e2e

import (
	"context"
	"math/big"
	"strings"
	"testing"

	"github.com/gateway-fm/cdk-erigon-lib/common"
	"github.com/holiman/uint256"
	ethereum "github.com/ledgerwatch/erigon"
	"github.com/ledgerwatch/erigon/core/types"
	"github.com/ledgerwatch/erigon/crypto"
	"github.com/ledgerwatch/erigon/ethclient"
	"github.com/ledgerwatch/erigon/test/operations"
	"github.com/ledgerwatch/erigon/zkevm/encoding"
	"github.com/ledgerwatch/erigon/zkevm/log"
	"github.com/stretchr/testify/require"
)

const (
	blockAddress    = "0xdD2FD4581271e230360230F9337D5c0430Bf44C0"
	blockPrivateKey = "0xde9be858da4a475276426320d5e9262ecfc3ba460bfac56360bfa6c4c28b4ee0"

	testVerified = false
)

func TestGetBatchSealTime(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	// latest batch seal time
	batchNum, err := operations.GetBatchNumber()
	require.NoError(t, err)
	batchSealTime, err := operations.GetBatchSealTime(new(big.Int).SetUint64(batchNum))
	require.Equal(t, batchSealTime, uint64(0))
	log.Infof("Batch number: %d", batchNum)

	// old batch seal time
	batchNum = batchNum - 1
	batch, err := operations.GetBatchByNumber(new(big.Int).SetUint64(batchNum))
	var maxTime uint64
	for _, block := range batch.Blocks {
		blockInfo, err := operations.GetBlockByHash(common.HexToHash(block.(string)))
		require.NoError(t, err)
		log.Infof("Block Timestamp: %+v", blockInfo.Timestamp)
		blockTime := uint64(blockInfo.Timestamp)
		if blockTime > maxTime {
			maxTime = blockTime
		}
	}
	batchSealTime, err = operations.GetBatchSealTime(new(big.Int).SetUint64(batchNum))
	require.NoError(t, err)
	log.Infof("Max block time: %d, batchSealTime: %d", maxTime, batchSealTime)
	require.Equal(t, maxTime, batchSealTime)
}

func TestClaimTx(t *testing.T) {
	ctx := context.Background()
	client, err := ethclient.Dial(operations.DefaultL2NetworkURL)
	transToken(t, ctx, client, uint256.NewInt(encoding.Gwei), operations.DefaultSequencerAddress)

	from := common.HexToAddress(operations.DefaultSequencerAddress)
	to := common.HexToAddress(operations.DefaultL2AdminAddress)
	nonce, err := client.PendingNonceAt(ctx, from)
	gas, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From:  from,
		To:    &to,
		Value: uint256.NewInt(10),
	})
	require.NoError(t, err)
	var tx types.Transaction = &types.LegacyTx{
		CommonTx: types.CommonTx{
			Nonce: nonce,
			To:    &to,
			Gas:   gas,
			Value: uint256.NewInt(10),
		},
		GasPrice: uint256.MustFromBig(big.NewInt(0)),
	}

	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(operations.DefaultSequencerPrivateKey, "0x"))
	require.NoError(t, err)

	signer := types.MakeSigner(operations.GetTestChainConfig(operations.DefaultL2ChainID), 1)
	signedTx, err := types.SignTx(tx, *signer, privateKey)
	require.NoError(t, err)

	err = client.SendTransaction(ctx, signedTx)
	require.NoError(t, err)

	err = operations.WaitTxToBeMined(ctx, client, signedTx, operations.DefaultTimeoutTxToBeMined)
	require.NoError(t, err)
}

func TestWhiteAndBlockList(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := context.Background()
	client, err := ethclient.Dial(operations.DefaultL2NetworkURL)
	from := common.HexToAddress(operations.DefaultL2AdminAddress)
	to := common.HexToAddress(blockAddress)
	nonce, err := client.PendingNonceAt(ctx, from)
	gasPrice, err := client.SuggestGasPrice(ctx)
	gas, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From:  from,
		To:    &to,
		Value: uint256.NewInt(10),
	})
	require.NoError(t, err)
	var tx types.Transaction = &types.LegacyTx{
		CommonTx: types.CommonTx{
			Nonce: nonce,
			To:    &to,
			Gas:   gas,
			Value: uint256.NewInt(10),
		},
		GasPrice: uint256.MustFromBig(gasPrice),
	}

	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(operations.DefaultL2AdminPrivateKey, "0x"))
	require.NoError(t, err)

	signer := types.MakeSigner(operations.GetTestChainConfig(operations.DefaultL2ChainID), 1)
	signedTx, err := types.SignTx(tx, *signer, privateKey)
	require.NoError(t, err)

	err = client.SendTransaction(ctx, signedTx)
	log.Infof("err:%v", err)
	require.True(t, strings.Contains(err.Error(), "INTERNAL_ERROR: blocked receiver"))
}

func TestRPCAPI(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	var err error
	for i := 0; i < 1000; i++ {
		_, err1 := operations.GetEthSyncing(operations.DefaultL2NetworkURL)
		if err1 != nil {
			err = err1
			break
		}
	}
	require.True(t, strings.Contains(err.Error(), "rate limit exceeded"))

	//for i := 0; i < 1000; i++ {
	//	_, err1 := operations.GetEthSyncing(operations.DefaultL2NetworkURL + "/apikey1")
	//	require.NoError(t, err1)
	//}
}

func TestChainID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	chainID, err := operations.GetNetVersion(operations.DefaultL2NetworkURL)
	require.NoError(t, err)
	require.Equal(t, chainID, operations.DefaultL2ChainID)
}

func TestInnerTx(t *testing.T) {
	ctx := context.Background()
	client, err := ethclient.Dial(operations.DefaultL2NetworkURL)
	require.NoError(t, err)
	txHash := transToken(t, ctx, client, uint256.NewInt(encoding.Gwei), operations.DefaultSequencerAddress)
	log.Infof("txHash: %s", txHash)

	result, err := operations.GetInternalTransactions(common.HexToHash(txHash))
	require.NoError(t, err)
	require.Greater(t, len(result), 0)
	require.Equal(t, result[0].From, operations.DefaultL2AdminAddress)

	tx, err := operations.GetTransactionByHash(common.HexToHash(txHash))
	require.NoError(t, err)
	log.Infof("tx: %+v", tx.BlockNumber)
	result1, err := operations.GetBlockInternalTransactions(new(big.Int).SetUint64(uint64(*tx.BlockNumber)))
	require.NoError(t, err)
	require.Greater(t, len(result1), 0)
	require.Equal(t, result1[common.HexToHash(txHash)][0].From, operations.DefaultL2AdminAddress)
}

func TestEthTransfer(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	if !testVerified {
		return
	}

	ctx := context.Background()
	auth, err := operations.GetAuth(operations.DefaultL2AdminPrivateKey, operations.DefaultL2ChainID)
	require.NoError(t, err)
	client, err := ethclient.Dial(operations.DefaultL2NetworkURL)
	require.NoError(t, err)

	from := common.HexToAddress(operations.DefaultL2AdminAddress)
	to := common.HexToAddress(operations.DefaultSequencerAddress)
	nonce, err := client.PendingNonceAt(ctx, from)
	require.NoError(t, err)
	var tx types.Transaction = &types.LegacyTx{
		CommonTx: types.CommonTx{
			Nonce: nonce,
			To:    &to,
			Gas:   21000,
			Value: uint256.NewInt(0),
		},
		GasPrice: uint256.NewInt(10 * encoding.Gwei),
	}
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(operations.DefaultL2AdminPrivateKey, "0x"))
	require.NoError(t, err)
	signer := types.MakeSigner(operations.GetTestChainConfig(operations.DefaultL2ChainID), 1)
	signedTx, err := types.SignTx(tx, *signer, privateKey)
	var txs []*types.Transaction
	txs = append(txs, &signedTx)
	_, err = operations.ApplyL2Txs(ctx, txs, auth, client, operations.VerifiedConfirmationLevel)
	require.NoError(t, err)
}

func TestGasPrice(t *testing.T) {
	ctx := context.Background()
	client, err := ethclient.Dial(operations.DefaultL2NetworkURL)
	log.Infof("Start TestGasPrice")
	gasPrice1, err := operations.GetGasPrice()
	gasPrice2 := gasPrice1
	require.NoError(t, err)
	for i := 1; i < 10; i++ {
		temp, err := operations.GetGasPrice()
		require.NoError(t, err)
		if temp > gasPrice2 {
			gasPrice2 = temp
		}
		require.NoError(t, err)

		from := common.HexToAddress(operations.DefaultL2AdminAddress)
		to := common.HexToAddress(operations.DefaultSequencerAddress)
		nonce, err := client.PendingNonceAt(ctx, from)
		require.NoError(t, err)
		var tx types.Transaction = &types.LegacyTx{
			CommonTx: types.CommonTx{
				Nonce: nonce,
				To:    &to,
				Gas:   21000,
				Value: uint256.NewInt(0),
			},
			GasPrice: uint256.NewInt(uint64(i) * 200 * encoding.Gwei),
		}
		privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(operations.DefaultL2AdminPrivateKey, "0x"))
		require.NoError(t, err)
		signer := types.MakeSigner(operations.GetTestChainConfig(operations.DefaultL2ChainID), 1)
		signedTx, err := types.SignTx(tx, *signer, privateKey)
		require.NoError(t, err)
		log.Infof("GP:%v", tx.GetPrice())
		err = client.SendTransaction(ctx, signedTx)
		err = operations.WaitTxToBeMined(ctx, client, signedTx, operations.DefaultTimeoutTxToBeMined)
		require.NoError(t, err)
	}
	require.NoError(t, err)
	log.Infof("gasPrice: [%d,%d]", gasPrice1, gasPrice2)
	require.Greater(t, gasPrice2, gasPrice1)
}

func transToken(t *testing.T, ctx context.Context, client *ethclient.Client, amount *uint256.Int, toAddress string) string {
	auth, err := operations.GetAuth(operations.DefaultL2AdminPrivateKey, operations.DefaultL2ChainID)
	nonce, err := client.PendingNonceAt(ctx, auth.From)
	gasPrice, err := client.SuggestGasPrice(ctx)
	require.NoError(t, err)

	to := common.HexToAddress(toAddress)
	gas, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From:  auth.From,
		To:    &to,
		Value: amount,
	})
	require.NoError(t, err)

	var tx types.Transaction = &types.LegacyTx{
		CommonTx: types.CommonTx{
			Nonce: nonce,
			To:    &to,
			Gas:   gas,
			Value: amount,
		},
		GasPrice: uint256.MustFromBig(gasPrice),
	}

	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(operations.DefaultL2AdminPrivateKey, "0x"))
	require.NoError(t, err)

	signer := types.MakeSigner(operations.GetTestChainConfig(operations.DefaultL2ChainID), 1)
	signedTx, err := types.SignTx(tx, *signer, privateKey)
	require.NoError(t, err)

	err = client.SendTransaction(ctx, signedTx)
	require.NoError(t, err)

	err = operations.WaitTxToBeMined(ctx, client, signedTx, operations.DefaultTimeoutTxToBeMined)
	require.NoError(t, err)

	return signedTx.Hash().String()
}

func TestMinGasPrice(t *testing.T) {
	ctx := context.Background()
	client, err := ethclient.Dial(operations.DefaultL2NetworkURL)
	log.Infof("Start TestMinGasPrice")
	require.NoError(t, err)
	for i := 1; i < 3; i++ {
		temp, err := operations.GetMinGasPrice()
		log.Infof("minGP: [%d]", temp)
		if temp > 1 {
			temp = temp - 1
		}
		require.NoError(t, err)

		from := common.HexToAddress(operations.DefaultL2AdminAddress)
		to := common.HexToAddress(operations.DefaultSequencerAddress)
		nonce, err := client.PendingNonceAt(ctx, from)
		require.NoError(t, err)
		var tx types.Transaction = &types.LegacyTx{
			CommonTx: types.CommonTx{
				Nonce: nonce,
				To:    &to,
				Gas:   21000,
				Value: uint256.NewInt(0),
			},
			GasPrice: uint256.NewInt(temp),
		}
		privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(operations.DefaultL2AdminPrivateKey, "0x"))
		require.NoError(t, err)
		signer := types.MakeSigner(operations.GetTestChainConfig(operations.DefaultL2ChainID), 1)
		signedTx, err := types.SignTx(tx, *signer, privateKey)
		require.NoError(t, err)
		log.Infof("GP:%v", tx.GetPrice())
		err = client.SendTransaction(ctx, signedTx)
		require.Error(t, err)
	}
	for i := 3; i < 5; i++ {
		temp, err := operations.GetMinGasPrice()
		log.Infof("minGP: [%d]", temp)
		require.NoError(t, err)

		from := common.HexToAddress(operations.DefaultL2AdminAddress)
		to := common.HexToAddress(operations.DefaultSequencerAddress)
		nonce, err := client.PendingNonceAt(ctx, from)
		require.NoError(t, err)
		var tx types.Transaction = &types.LegacyTx{
			CommonTx: types.CommonTx{
				Nonce: nonce,
				To:    &to,
				Gas:   21000,
				Value: uint256.NewInt(0),
			},
			GasPrice: uint256.NewInt(temp),
		}
		privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(operations.DefaultL2AdminPrivateKey, "0x"))
		require.NoError(t, err)
		signer := types.MakeSigner(operations.GetTestChainConfig(operations.DefaultL2ChainID), 1)
		signedTx, err := types.SignTx(tx, *signer, privateKey)
		require.NoError(t, err)
		log.Infof("GP:%v", tx.GetPrice())
		err = client.SendTransaction(ctx, signedTx)
		require.NoError(t, err)
	}
	require.NoError(t, err)
}
