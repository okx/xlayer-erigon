package apollo

import (
	"flag"
	"fmt"
	"math"
	"strings"

	libcommon "github.com/gateway-fm/cdk-erigon-lib/common"
	"github.com/ledgerwatch/erigon/cmd/utils"
	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/urfave/cli/v2"
)

func createMockContext(flags []cli.Flag) *cli.Context {
	set := flag.NewFlagSet("", flag.ContinueOnError)
	for _, f := range flags {
		f.Apply(set)
	}

	context := cli.NewContext(nil, set, nil)
	return context
}

// loadZkConfig loads the generic zkEVM eth apollo configurations
func loadZkConfig(ctx *cli.Context, ethCfg *ethconfig.Config) {
	if ethCfg.Zk == nil {
		ethCfg.Zk = &ethconfig.Zk{}
	}

	if ctx.IsSet(utils.L2ChainIdFlag.Name) {
		ethCfg.Zk.L2ChainId = ctx.Uint64(utils.L2ChainIdFlag.Name)
	}
	if ctx.IsSet(utils.L2RpcUrlFlag.Name) {
		ethCfg.Zk.L2RpcUrl = ctx.String(utils.L2RpcUrlFlag.Name)
	}
	if ctx.IsSet(utils.L2DataStreamerUrlFlag.Name) {
		ethCfg.Zk.L2DataStreamerUrl = ctx.String(utils.L2DataStreamerUrlFlag.Name)
	}
	if ctx.IsSet(utils.L2DataStreamerTimeout.Name) {
		ethCfg.Zk.L2DataStreamerTimeout = ctx.Duration(utils.L2DataStreamerTimeout.Name)
	}
	if ctx.IsSet(utils.L1SyncStartBlock.Name) {
		ethCfg.Zk.L1SyncStartBlock = ctx.Uint64(utils.L1SyncStartBlock.Name)
	}
	if ctx.IsSet(utils.L1SyncStopBatch.Name) {
		ethCfg.Zk.L1SyncStopBatch = ctx.Uint64(utils.L1SyncStopBatch.Name)
	}
	if ctx.IsSet(utils.L1ChainIdFlag.Name) {
		ethCfg.Zk.L1ChainId = ctx.Uint64(utils.L1ChainIdFlag.Name)
	}
	if ctx.IsSet(utils.L1RpcUrlFlag.Name) {
		ethCfg.Zk.L1RpcUrl = ctx.String(utils.L1RpcUrlFlag.Name)
	}
	if ctx.IsSet(utils.AddressSequencerFlag.Name) {
		ethCfg.Zk.AddressSequencer = libcommon.HexToAddress(ctx.String(utils.AddressSequencerFlag.Name))
	}
	if ctx.IsSet(utils.AddressAdminFlag.Name) {
		ethCfg.Zk.AddressAdmin = libcommon.HexToAddress(ctx.String(utils.AddressAdminFlag.Name))
	}
	if ctx.IsSet(utils.AddressRollupFlag.Name) {
		ethCfg.Zk.AddressRollup = libcommon.HexToAddress(ctx.String(utils.AddressRollupFlag.Name))
	}
	if ctx.IsSet(utils.AddressZkevmFlag.Name) {
		ethCfg.Zk.AddressZkevm = libcommon.HexToAddress(ctx.String(utils.AddressZkevmFlag.Name))
	}
	if ctx.IsSet(utils.AddressGerManagerFlag.Name) {
		ethCfg.Zk.AddressGerManager = libcommon.HexToAddress(ctx.String(utils.AddressGerManagerFlag.Name))
	}
	if ctx.IsSet(utils.L1RollupIdFlag.Name) {
		ethCfg.Zk.L1RollupId = ctx.Uint64(utils.L1RollupIdFlag.Name)
	}
	if ctx.IsSet(utils.L1BlockRangeFlag.Name) {
		ethCfg.Zk.L1BlockRange = ctx.Uint64(utils.L1BlockRangeFlag.Name)
	}
	if ctx.IsSet(utils.L1QueryDelayFlag.Name) {
		ethCfg.Zk.L1QueryDelay = ctx.Uint64(utils.L1QueryDelayFlag.Name)
	}
	if ctx.IsSet(utils.L1MaticContractAddressFlag.Name) {
		ethCfg.Zk.L1MaticContractAddress = libcommon.HexToAddress(ctx.String(utils.L1MaticContractAddressFlag.Name))
	}
	if ctx.IsSet(utils.L1FirstBlockFlag.Name) {
		ethCfg.Zk.L1FirstBlock = ctx.Uint64(utils.L1FirstBlockFlag.Name)
	}
	if ctx.IsSet(utils.RpcRateLimitsFlag.Name) {
		ethCfg.Zk.RpcRateLimits = ctx.Int(utils.RpcRateLimitsFlag.Name)
	}
	if ctx.IsSet(utils.DatastreamVersionFlag.Name) {
		ethCfg.Zk.DatastreamVersion = ctx.Int(utils.DatastreamVersionFlag.Name)
	}
	if ctx.IsSet(utils.RebuildTreeAfterFlag.Name) {
		ethCfg.Zk.RebuildTreeAfter = ctx.Uint64(utils.RebuildTreeAfterFlag.Name)
	}
	if ctx.IsSet(utils.ExecutorUrls.Name) {
		ethCfg.Zk.ExecutorUrls = strings.Split(ctx.String(utils.ExecutorUrls.Name), ",")
	}
	if ctx.IsSet(utils.ExecutorStrictMode.Name) {
		ethCfg.Zk.ExecutorStrictMode = ctx.Bool(utils.ExecutorStrictMode.Name)
	}
	if ctx.IsSet(utils.ExecutorRequestTimeout.Name) {
		ethCfg.Zk.ExecutorRequestTimeout = ctx.Duration(utils.ExecutorRequestTimeout.Name)
	}
	if ctx.IsSet(utils.ExecutorMaxConcurrentRequests.Name) {
		ethCfg.Zk.ExecutorMaxConcurrentRequests = ctx.Int(utils.ExecutorMaxConcurrentRequests.Name)
	}
	if ctx.IsSet(utils.AllowFreeTransactions.Name) {
		ethCfg.Zk.AllowFreeTransactions = ctx.Bool(utils.AllowFreeTransactions.Name)
	}
	if ctx.IsSet(utils.AllowPreEIP155Transactions.Name) {
		ethCfg.Zk.AllowPreEIP155Transactions = ctx.Bool(utils.AllowPreEIP155Transactions.Name)
	}
	if ctx.IsSet(utils.EffectiveGasPriceForEthTransfer.Name) {
		effectiveGasPriceForEthTransferVal := ctx.Float64(utils.EffectiveGasPriceForEthTransfer.Name)
		effectiveGasPriceForEthTransferVal = math.Max(effectiveGasPriceForEthTransferVal, 0)
		effectiveGasPriceForEthTransferVal = math.Min(effectiveGasPriceForEthTransferVal, 1)
		ethCfg.Zk.EffectiveGasPriceForEthTransfer = uint8(math.Round(effectiveGasPriceForEthTransferVal * 255.0))
	}
	if ctx.IsSet(utils.EffectiveGasPriceForErc20Transfer.Name) {
		effectiveGasPriceForErc20TransferVal := ctx.Float64(utils.EffectiveGasPriceForErc20Transfer.Name)
		effectiveGasPriceForErc20TransferVal = math.Max(effectiveGasPriceForErc20TransferVal, 0)
		effectiveGasPriceForErc20TransferVal = math.Min(effectiveGasPriceForErc20TransferVal, 1)
		ethCfg.Zk.EffectiveGasPriceForErc20Transfer = uint8(math.Round(effectiveGasPriceForErc20TransferVal * 255.0))
	}
	if ctx.IsSet(utils.EffectiveGasPriceForContractInvocation.Name) {
		effectiveGasPriceForContractInvocationVal := ctx.Float64(utils.EffectiveGasPriceForContractInvocation.Name)
		effectiveGasPriceForContractInvocationVal = math.Max(effectiveGasPriceForContractInvocationVal, 0)
		effectiveGasPriceForContractInvocationVal = math.Min(effectiveGasPriceForContractInvocationVal, 1)
		ethCfg.Zk.EffectiveGasPriceForContractInvocation = uint8(math.Round(effectiveGasPriceForContractInvocationVal * 255.0))
	}
	if ctx.IsSet(utils.EffectiveGasPriceForContractDeployment.Name) {
		effectiveGasPriceForContractDeploymentVal := ctx.Float64(utils.EffectiveGasPriceForContractDeployment.Name)
		effectiveGasPriceForContractDeploymentVal = math.Max(effectiveGasPriceForContractDeploymentVal, 0)
		effectiveGasPriceForContractDeploymentVal = math.Min(effectiveGasPriceForContractDeploymentVal, 1)
		ethCfg.Zk.EffectiveGasPriceForContractDeployment = uint8(math.Round(effectiveGasPriceForContractDeploymentVal * 255.0))
	}
	if ctx.IsSet(utils.DefaultGasPrice.Name) {
		ethCfg.Zk.DefaultGasPrice = ctx.Uint64(utils.DefaultGasPrice.Name)
	}
	if ctx.IsSet(utils.MaxGasPrice.Name) {
		ethCfg.Zk.MaxGasPrice = ctx.Uint64(utils.MaxGasPrice.Name)
	}
	if ctx.IsSet(utils.GasPriceFactor.Name) {
		ethCfg.Zk.GasPriceFactor = ctx.Float64(utils.GasPriceFactor.Name)
	}
	if ctx.IsSet(utils.WitnessFullFlag.Name) {
		ethCfg.Zk.WitnessFull = ctx.Bool(utils.WitnessFullFlag.Name)
	}
	if ctx.IsSet(utils.SyncLimit.Name) {
		ethCfg.Zk.SyncLimit = ctx.Uint64(utils.SyncLimit.Name)
	}
	if ctx.IsSet(utils.SupportGasless.Name) {
		ethCfg.Zk.Gasless = ctx.Bool(utils.SupportGasless.Name)
	}
	if ctx.IsSet(utils.DebugNoSync.Name) {
		ethCfg.Zk.DebugNoSync = ctx.Bool(utils.DebugNoSync.Name)
	}
	if ctx.IsSet(utils.DebugLimit.Name) {
		ethCfg.Zk.DebugLimit = ctx.Uint64(utils.DebugLimit.Name)
	}
	if ctx.IsSet(utils.DebugStep.Name) {
		ethCfg.Zk.DebugStep = ctx.Uint64(utils.DebugStep.Name)
	}
	if ctx.IsSet(utils.DebugStepAfter.Name) {
		ethCfg.Zk.DebugStepAfter = ctx.Uint64(utils.DebugStepAfter.Name)
	}
	if ctx.IsSet(utils.PoolManagerUrl.Name) {
		ethCfg.Zk.PoolManagerUrl = ctx.String(utils.PoolManagerUrl.Name)
	}
	if ctx.IsSet(utils.DisableVirtualCounters.Name) {
		ethCfg.Zk.DisableVirtualCounters = ctx.Bool(utils.DisableVirtualCounters.Name)
	}
	if ctx.IsSet(utils.ExecutorPayloadOutput.Name) {
		ethCfg.Zk.ExecutorPayloadOutput = ctx.String(utils.ExecutorPayloadOutput.Name)
	}
	// X Layer configs. Do not set nacos config as it is read from env
	if ctx.IsSet(utils.AllowInternalTransactions.Name) {
		ethCfg.Zk.XLayer.EnableInnerTx = ctx.Bool(utils.AllowInternalTransactions.Name)
	}
}

func getNamespacePrefix(namespace string) (string, error) {
	items := strings.Split(namespace, "-")
	if len(items) < NamespaceSplits {
		return "", fmt.Errorf("invalid namespace: %s, no separator \"-\" present, please configure apollo namespace in the correct format \"prefix-item\"", namespace)
	}
	return items[0], nil
}

func getNamespaceSuffix(namespace string) (string, error) {
	items := strings.Split(namespace, "-")
	if len(items) < NamespaceSplits {
		return "", fmt.Errorf("invalid namespace: %s, no separator \"-\" present, please configure apollo namespace in the correct format \"item-suffix\"", namespace)
	}
	return items[len(items)-1], nil
}
