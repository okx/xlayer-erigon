package e2e

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/holiman/uint256"
	ethereum "github.com/ledgerwatch/erigon"
	"github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon/accounts/abi/bind"
	"github.com/ledgerwatch/erigon/core/types"
	"github.com/ledgerwatch/erigon/crypto"
	"github.com/ledgerwatch/erigon/ethclient"
	"gopkg.in/yaml.v2"

	"github.com/ledgerwatch/erigon/test/operations"
	"github.com/ledgerwatch/erigon/zkevm/encoding"
	"github.com/ledgerwatch/erigon/zkevm/etherman/smartcontracts/polygonzkevmbridge"
	"github.com/ledgerwatch/erigon/zkevm/log"

	"github.com/stretchr/testify/require"
)

const (
	blockAddress    = "0xdD2FD4581271e230360230F9337D5c0430Bf44C0"
	blockPrivateKey = "0xde9be858da4a475276426320d5e9262ecfc3ba460bfac56360bfa6c4c28b4ee0"

	testVerified = true
)

func TestGetBatchSealTime(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	// latest batch seal time
	var batchNum uint64
	var batchSealTime uint64
	var err error
	for i := 0; i < 50; i++ {
		batchNum, err = operations.GetBatchNumber()
		require.NoError(t, err)
		batchSealTime, err = operations.GetBatchSealTime(new(big.Int).SetUint64(batchNum))
		require.Equal(t, batchSealTime, uint64(0))
		log.Infof("Batch number: %d, times:%v", batchNum, i)
		if batchNum > 1 {
			break
		}
		time.Sleep(1 * time.Second)
	}

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

func TestBridgeTx(t *testing.T) {
	ctx := context.Background()
	l1Client, err := ethclient.Dial(operations.DefaultL1NetworkURL)
	require.NoError(t, err)
	l2Client, err := ethclient.Dial(operations.DefaultL2NetworkURL)
	require.NoError(t, err)
	transToken(t, ctx, l2Client, uint256.NewInt(encoding.Gwei), operations.DefaultL2AdminAddress)

	amount := new(big.Int).SetUint64(100)
	//layer2 network id
	var destNetwork uint32 = 1
	destAddr := common.HexToAddress(operations.DefaultL2AdminAddress)
	auth, err := operations.GetAuth(operations.DefaultL1AdminPrivateKey, operations.DefaultL1ChainID)
	require.NoError(t, err)

	wethAddress := common.HexToAddress("0x17a2a2e444a7f3446877d1b71eaa2b2ae7533baf")
	wethToken, err := operations.NewToken(wethAddress, l2Client)
	require.NoError(t, err)

	balanceBefore, err := wethToken.BalanceOf(&bind.CallOpts{}, destAddr)
	require.NoError(t, err)

	err = sendBridgeAsset(ctx, common.Address{}, amount, destNetwork, &destAddr, []byte{}, auth, common.HexToAddress(operations.BridgeAddr), l1Client)
	require.NoError(t, err)

	const maxAttempts = 120

	var balanceAfter *big.Int
	for i := 0; i < maxAttempts; i++ {
		time.Sleep(1 * time.Second)

		balanceAfter, err = wethToken.BalanceOf(&bind.CallOpts{}, destAddr)
		require.NoError(t, err)

		if balanceAfter.Cmp(balanceBefore) > 0 {
			return
		}
	}

	t.Errorf("bridge transaction failed after %d seconds: balance did not increase (before: %s, after: %s)",
		maxAttempts,
		balanceBefore.String(),
		balanceAfter.String(),
	)
}

func TestClaimTx(t *testing.T) {
	ctx := context.Background()
	client, err := ethclient.Dial(operations.DefaultL2NetworkURL)
	transToken(t, ctx, client, uint256.NewInt(encoding.Gwei), operations.DefaultL2AdminAddress)

	from := common.HexToAddress(operations.DefaultL2AdminAddress)
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

	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(operations.DefaultL2AdminPrivateKey, "0x"))
	require.NoError(t, err)

	signer := types.MakeSigner(operations.GetTestChainConfig(operations.DefaultL2ChainID), 1, 0)
	signedTx, err := types.SignTx(tx, *signer, privateKey)
	require.NoError(t, err)

	err = client.SendTransaction(ctx, signedTx)
	require.NoError(t, err)

	err = operations.WaitTxToBeMined(ctx, client, signedTx, operations.DefaultTimeoutTxToBeMined)
	require.NoError(t, err)
}

func TestNewAccFreeGas(t *testing.T) {
	ctx := context.Background()
	client, _ := ethclient.Dial(operations.DefaultL2NetworkURL)
	transToken(t, ctx, client, uint256.NewInt(encoding.Gwei), operations.DefaultL2AdminAddress)
	var gas uint64 = 21000

	//newAcc transfer failed
	from := common.HexToAddress(operations.DefaultL2NewAcc1Address)
	to := common.HexToAddress(operations.DefaultL2AdminAddress)
	nonce, err := client.PendingNonceAt(ctx, from)
	require.NoError(t, err)
	var tx types.Transaction = &types.LegacyTx{
		CommonTx: types.CommonTx{
			Nonce: nonce,
			To:    &to,
			Gas:   gas,
			Value: uint256.NewInt(0),
		},
		GasPrice: uint256.MustFromBig(big.NewInt(0)),
	}
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(operations.DefaultL2NewAcc1PrivateKey, "0x"))
	require.NoError(t, err)
	signer := types.MakeSigner(operations.GetTestChainConfig(operations.DefaultL2ChainID), 1, 0)
	signedTx, err := types.SignTx(tx, *signer, privateKey)
	require.NoError(t, err)
	err = client.SendTransaction(ctx, signedTx)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "RPC error response: FEE_TOO_LOW: underpriced"), "Expected error message not found")
	err = operations.WaitTxToBeMined(ctx, client, signedTx, 5)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "context deadline exceeded"), "Expected error message not found")

	// seq -> newAcc
	from = common.HexToAddress(operations.DefaultL2AdminAddress)
	to = common.HexToAddress(operations.DefaultL2NewAcc1Address)
	nonce, err = client.PendingNonceAt(ctx, from)
	require.NoError(t, err)
	tx = &types.LegacyTx{
		CommonTx: types.CommonTx{
			Nonce: nonce,
			To:    &to,
			Gas:   gas,
			Value: uint256.NewInt(10),
		},
		GasPrice: uint256.MustFromBig(big.NewInt(0)),
	}
	privateKey, err = crypto.HexToECDSA(strings.TrimPrefix(operations.DefaultL1AdminPrivateKey, "0x"))
	require.NoError(t, err)
	signedTx, err = types.SignTx(tx, *signer, privateKey)
	require.NoError(t, err)
	err = client.SendTransaction(ctx, signedTx)
	require.NoError(t, err)
	err = operations.WaitTxToBeMined(ctx, client, signedTx, operations.DefaultTimeoutTxToBeMined)
	require.NoError(t, err)

	// newAcc transfer success
	from = common.HexToAddress(operations.DefaultL2NewAcc1Address)
	to = common.HexToAddress(operations.DefaultL2AdminAddress)
	nonce, err = client.PendingNonceAt(ctx, from)
	require.NoError(t, err)
	tx = &types.LegacyTx{
		CommonTx: types.CommonTx{
			Nonce: nonce,
			To:    &to,
			Gas:   gas,
			Value: uint256.NewInt(0),
		},
		GasPrice: uint256.MustFromBig(big.NewInt(0)),
	}
	privateKey, err = crypto.HexToECDSA(strings.TrimPrefix(operations.DefaultL2NewAcc1PrivateKey, "0x"))
	require.NoError(t, err)
	signedTx, err = types.SignTx(tx, *signer, privateKey)
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
	require.NoError(t, err)

	from := common.HexToAddress(operations.DefaultL2AdminAddress)
	blockAddressConverted := common.HexToAddress(blockAddress)
	nonBlockAddress := common.HexToAddress(operations.DefaultL2NewAcc1Address)

	nonce, err := client.PendingNonceAt(ctx, from)
	require.NoError(t, err)

	gasPrice, err := client.SuggestGasPrice(ctx)
	require.NoError(t, err)

	gas, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From:  from,
		To:    &blockAddressConverted,
		Value: uint256.NewInt(10),
	})
	require.NoError(t, err)

	var txToBlockAddress types.Transaction = &types.LegacyTx{
		CommonTx: types.CommonTx{
			Nonce: nonce,
			To:    &blockAddressConverted,
			Gas:   gas,
			Value: uint256.NewInt(10),
		},
		GasPrice: uint256.MustFromBig(gasPrice),
	}

	var txToNonBlockAddress types.Transaction = &types.LegacyTx{
		CommonTx: types.CommonTx{
			Nonce: nonce,
			To:    &nonBlockAddress,
			Gas:   gas,
			Value: uint256.NewInt(10),
		},
		GasPrice: uint256.MustFromBig(gasPrice),
	}

	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(operations.DefaultL2AdminPrivateKey, "0x"))
	require.NoError(t, err)

	signer := types.MakeSigner(operations.GetTestChainConfig(operations.DefaultL2ChainID), 1, 0)

	signedTxToBlockAddress, err := types.SignTx(txToBlockAddress, *signer, privateKey)
	require.NoError(t, err)

	err = client.SendTransaction(ctx, signedTxToBlockAddress)
	log.Infof("err:%v", err)
	require.True(t, strings.Contains(err.Error(), "INTERNAL_ERROR: blocked receiver"))

	signedTxToNonBlockAddress, err := types.SignTx(txToNonBlockAddress, *signer, privateKey)
	require.NoError(t, err)

	err = client.SendTransaction(ctx, signedTxToNonBlockAddress)
	require.NoError(t, err)

	//TODO: sender in blocklist should fail
	//now only admin account have balance. So we may add another account that has balance.
}

func TestRPCAPI(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	config, err := LoadConfig("../../test/config/test.erigon.rpc.config.yaml")
	require.NoError(t, err)

	if config.HTTPAPIKeys != "" {

		_, err := operations.GetEthSyncing(operations.DefaultL2NetworkURL)
		require.Error(t, err)
		require.True(t, strings.Contains(err.Error(), "no authentication"))

		_, err = operations.GetEthSyncing(operations.DefaultL2NetworkURL + "/45543e0adc5dd3e316044909d32501a5")
		require.NoError(t, err)
	} else {

		var rateErr error
		for i := 0; i < 1000; i++ {
			_, err1 := operations.GetEthSyncing(operations.DefaultL2NetworkURL)
			if err1 != nil {
				rateErr = err1
				break
			}
		}

		require.True(t, strings.Contains(rateErr.Error(), "rate limit exceeded"))
	}
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
	txHash := transToken(t, ctx, client, uint256.NewInt(encoding.Gwei), operations.DefaultL2AdminAddress)
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
	to := common.HexToAddress(operations.DefaultL2NewAcc1Address)
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
	signer := types.MakeSigner(operations.GetTestChainConfig(operations.DefaultL2ChainID), 1, 0)
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
	for i := 1; i < 100; i++ {
		temp, err := operations.GetGasPrice()
		require.NoError(t, err)
		if temp > gasPrice2 {
			gasPrice2 = temp
		}
		require.NoError(t, err)

		from := common.HexToAddress(operations.DefaultL2AdminAddress)
		to := common.HexToAddress(operations.DefaultL2NewAcc1Address)
		nonce, err := client.PendingNonceAt(ctx, from)
		require.NoError(t, err)
		var tx types.Transaction = &types.LegacyTx{
			CommonTx: types.CommonTx{
				Nonce: nonce,
				To:    &to,
				Gas:   21000,
				Value: uint256.NewInt(0),
			},
			GasPrice: uint256.NewInt(uint64(i) * 10 * encoding.Gwei),
		}
		privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(operations.DefaultL2AdminPrivateKey, "0x"))
		require.NoError(t, err)
		signer := types.MakeSigner(operations.GetTestChainConfig(operations.DefaultL2ChainID), 1, 0)
		signedTx, err := types.SignTx(tx, *signer, privateKey)
		require.NoError(t, err)
		log.Infof("Get new GP:%v, TXGP:%v", temp, tx.GetPrice())
		err = client.SendTransaction(ctx, signedTx)
		time.Sleep(500 * time.Millisecond)
		//err = operations.WaitTxToBeMined(ctx, client, signedTx, operations.DefaultTimeoutTxToBeMined)
		//require.NoError(t, err)
		if gasPrice2 > gasPrice1 {
			log.Infof("GP compare ok: [%d,%d]", gasPrice1, gasPrice2)
			break
		}
	}
	require.NoError(t, err)
	log.Infof("gasPrice: [%d,%d]", gasPrice1, gasPrice2)
	require.Greater(t, gasPrice2, gasPrice1)
}

func TestMetrics(t *testing.T) {
	result, err := operations.GetMetricsPrometheus()
	require.NoError(t, err)
	require.Equal(t, strings.Contains(result, "sequencer_batch_execute_time"), true)
	require.Equal(t, strings.Contains(result, "sequencer_pool_tx_count"), true)

	// TODO: enable this test after metrics are enabled
	//result, err = operations.GetMetrics()
	//require.NoError(t, err)
	//require.Equal(t, strings.Contains(result, "zkevm_getBatchWitness"), true)
	//require.Equal(t, strings.Contains(result, "eth_sendRawTransaction"), true)
	//require.Equal(t, strings.Contains(result, "eth_getTransactionCount"), true)
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
	log.Infof("gas: %d", gas)
	log.Infof("gasPrice: %d", gasPrice)

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

	signer := types.MakeSigner(operations.GetTestChainConfig(operations.DefaultL2ChainID), 1, 0)
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

		from := common.HexToAddress(operations.DefaultL2NewAcc2Address)
		to := common.HexToAddress(operations.DefaultL1AdminAddress)
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
		privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(operations.DefaultL2NewAcc2PrivateKey, "0x"))
		require.NoError(t, err)
		signer := types.MakeSigner(operations.GetTestChainConfig(operations.DefaultL2ChainID), 1, 0)
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
		to := common.HexToAddress(operations.DefaultL1AdminAddress)
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
		signer := types.MakeSigner(operations.GetTestChainConfig(operations.DefaultL2ChainID), 1, 0)
		signedTx, err := types.SignTx(tx, *signer, privateKey)
		require.NoError(t, err)
		log.Infof("GP:%v", tx.GetPrice())
		err = client.SendTransaction(ctx, signedTx)
		require.NoError(t, err)
	}
	require.NoError(t, err)
}

func sendBridgeAsset(
	ctx context.Context, tokenAddr common.Address, amount *big.Int, destNetwork uint32, destAddr *common.Address,
	metadata []byte, auth *bind.TransactOpts, bridgeSCAddr common.Address, c *ethclient.Client,
) error {
	emptyAddr := common.Address{}
	if tokenAddr == emptyAddr {
		auth.Value = amount
	}
	if destAddr == nil {
		destAddr = &auth.From
	}
	if len(bridgeSCAddr) == 0 {
		return fmt.Errorf("Bridge address error")
	}

	br, err := polygonzkevmbridge.NewPolygonzkevmbridge(bridgeSCAddr, c)
	if err != nil {
		return err
	}
	tx, err := br.BridgeAsset(auth, destNetwork, *destAddr, amount, tokenAddr, true, metadata)
	if err != nil {
		return err
	}
	// wait transfer to be included in a batch
	const txTimeout = 60 * time.Second
	return operations.WaitTxToBeMined(ctx, c, tx, txTimeout)
}

type Config struct {
	HTTPMethodRateLimit string `yaml:"http.methodratelimit"`
	HTTPAPIKeys         string `yaml:"http.apikeys"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
