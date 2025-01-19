package jsonrpc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/ledgerwatch/erigon/crypto"
	"golang.org/x/crypto/sha3"
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
	"github.com/ledgerwatch/log/v3"
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

var EntryPointAddress = "0x0FA5445ed0a18817f714D33d89F44eE4452aEa77"

// var chainID = 195
var TargetMethod = "handleOps"
var TargetMethod2 = "getSignature2Decode"

var HelpABIData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"sig\",\"type\":\"bytes\"}],\"name\":\"getSignature2Decode\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"}]",
}

var PayABIData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"},{\"internalType\":\"bytes\",\"name\":\"ret\",\"type\":\"bytes\"}],\"name\":\"DelegateAndRevert\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"opIndex\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"}],\"name\":\"FailedOp\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"opIndex\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"inner\",\"type\":\"bytes\"}],\"name\":\"FailedOpWithRevert\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"returnData\",\"type\":\"bytes\"}],\"name\":\"PostOpReverted\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ReentrancyGuardReentrantCall\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"SenderAddressResult\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"aggregator\",\"type\":\"address\"}],\"name\":\"SignatureValidationFailed\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"userOpHash\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"factory\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"paymaster\",\"type\":\"address\"}],\"name\":\"AccountDeployed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"BeforeExecution\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"totalDeposit\",\"type\":\"uint256\"}],\"name\":\"Deposited\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"userOpHash\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"revertReason\",\"type\":\"bytes\"}],\"name\":\"PostOpRevertReason\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"aggregator\",\"type\":\"address\"}],\"name\":\"SignatureAggregatorChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"totalStaked\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"unstakeDelaySec\",\"type\":\"uint256\"}],\"name\":\"StakeLocked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"withdrawTime\",\"type\":\"uint256\"}],\"name\":\"StakeUnlocked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"withdrawAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"StakeWithdrawn\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"userOpHash\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"paymaster\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"actualGasCost\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"actualGasUsed\",\"type\":\"uint256\"}],\"name\":\"UserOperationEvent\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"userOpHash\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"name\":\"UserOperationPrefundTooLow\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"userOpHash\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"revertReason\",\"type\":\"bytes\"}],\"name\":\"UserOperationRevertReason\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"withdrawAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Withdrawn\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"unstakeDelaySec\",\"type\":\"uint32\"}],\"name\":\"addStake\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"delegateAndRevert\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"depositTo\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"deposits\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"deposit\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"staked\",\"type\":\"bool\"},{\"internalType\":\"uint112\",\"name\":\"stake\",\"type\":\"uint112\"},{\"internalType\":\"uint32\",\"name\":\"unstakeDelaySec\",\"type\":\"uint32\"},{\"internalType\":\"uint48\",\"name\":\"withdrawTime\",\"type\":\"uint48\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"getDepositInfo\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"deposit\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"staked\",\"type\":\"bool\"},{\"internalType\":\"uint112\",\"name\":\"stake\",\"type\":\"uint112\"},{\"internalType\":\"uint32\",\"name\":\"unstakeDelaySec\",\"type\":\"uint32\"},{\"internalType\":\"uint48\",\"name\":\"withdrawTime\",\"type\":\"uint48\"}],\"internalType\":\"structIStakeManager.DepositInfo\",\"name\":\"info\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"uint192\",\"name\":\"key\",\"type\":\"uint192\"}],\"name\":\"getNonce\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"initCode\",\"type\":\"bytes\"}],\"name\":\"getSenderAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"initCode\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"callData\",\"type\":\"bytes\"},{\"internalType\":\"bytes32\",\"name\":\"accountGasLimits\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"preVerificationGas\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"gasFees\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"paymasterAndData\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"internalType\":\"structPackedUserOperation\",\"name\":\"userOp\",\"type\":\"tuple\"}],\"name\":\"getUserOpHash\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"initCode\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"callData\",\"type\":\"bytes\"},{\"internalType\":\"bytes32\",\"name\":\"accountGasLimits\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"preVerificationGas\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"gasFees\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"paymasterAndData\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"internalType\":\"structPackedUserOperation[]\",\"name\":\"userOps\",\"type\":\"tuple[]\"},{\"internalType\":\"contractIAggregator\",\"name\":\"aggregator\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"internalType\":\"structIEntryPoint.UserOpsPerAggregator[]\",\"name\":\"opsPerAggregator\",\"type\":\"tuple[]\"},{\"internalType\":\"addresspayable\",\"name\":\"beneficiary\",\"type\":\"address\"}],\"name\":\"handleAggregatedOps\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"initCode\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"callData\",\"type\":\"bytes\"},{\"internalType\":\"bytes32\",\"name\":\"accountGasLimits\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"preVerificationGas\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"gasFees\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"paymasterAndData\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"internalType\":\"structPackedUserOperation[]\",\"name\":\"ops\",\"type\":\"tuple[]\"},{\"internalType\":\"addresspayable\",\"name\":\"beneficiary\",\"type\":\"address\"}],\"name\":\"handleOps\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint192\",\"name\":\"key\",\"type\":\"uint192\"}],\"name\":\"incrementNonce\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"callData\",\"type\":\"bytes\"},{\"components\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"verificationGasLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"callGasLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"paymasterVerificationGasLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"paymasterPostOpGasLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"preVerificationGas\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"paymaster\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"maxFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxPriorityFeePerGas\",\"type\":\"uint256\"}],\"internalType\":\"structEntryPoint.MemoryUserOp\",\"name\":\"mUserOp\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"userOpHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"prefund\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"contextOffset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"preOpGas\",\"type\":\"uint256\"}],\"internalType\":\"structEntryPoint.UserOpInfo\",\"name\":\"opInfo\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"context\",\"type\":\"bytes\"}],\"name\":\"innerHandleOp\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"actualGasCost\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint192\",\"name\":\"\",\"type\":\"uint192\"}],\"name\":\"nonceSequenceNumber\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"unlockStake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"withdrawAddress\",\"type\":\"address\"}],\"name\":\"withdrawStake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"withdrawAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"withdrawAmount\",\"type\":\"uint256\"}],\"name\":\"withdrawTo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
}

// Hash calculates the Keccak256 hash of the user operation.
func (op *PackedUserOperation) Hash() ([]byte, error) {
	// ABI encoding of user operation fields

	Uint256, _ := abi.NewType("uint256", "", nil)
	Address, _ := abi.NewType("address", "", nil)
	Bytes, _ := abi.NewType("bytes", "", nil)
	Bytes32, _ := abi.NewType("bytes32", "", nil)
	arguments := abi.Arguments{
		{Type: Address}, // Sender
		{Type: Uint256}, // Nonce
		{Type: Bytes},   // InitCode
		{Type: Bytes},   // CallData
		{Type: Bytes32}, // AccountGasLimits
		{Type: Uint256}, // PreVerificationGas
		{Type: Bytes32}, // GasFees
		{Type: Bytes},   // PaymasterAndData
		{Type: Bytes},   // Signature
	}

	// Pack data into ABI-encoded bytes
	data, err := arguments.Pack(
		op.Sender,
		op.Nonce,
		op.InitCode,
		op.CallData,
		op.AccountGasLimits,
		op.PreVerificationGas,
		op.GasFees,
		op.PaymasterAndData,
		op.Signature,
	)
	if err != nil {
		log.Error("zjg, Failed to pack data:", err)
		return nil, err
	}

	// Hash the ABI-encoded data
	hash := sha3.NewLegacyKeccak256()
	hash.Write(data)
	return hash.Sum(nil), nil
}

func recoverSigner(signedHash []byte, signature []byte) (libcommon.Address, error) {
	// Decode the signature (v, r, s)
	if len(signature) != 65 {
		log.Error("zjg, Invalid signature length")
		return libcommon.Address{}, fmt.Errorf("invalid signature length")
	}

	// Signature format: 0x{r}{s}{v}
	r := signature[:32]
	s := signature[32:64]
	v := signature[64]

	// 如果 v 是 0 或 1，转换为 27 或 28
	if v != 27 && v != 28 {
		if v == 0 || v == 1 {
			v += 27
		} else {
			log.Error("zjg, Invalid signature v value")
			return libcommon.Address{}, errors.New("invalid signature v value")
		}
	}

	// Recover public key using ECDSA
	pubKey, err := crypto.SigToPub(signedHash, append(append(r, s...), v))
	if err != nil {
		log.Info(fmt.Sprintf("zjg, Failed to recover public key: %v", err))
		return libcommon.Address{}, err
	}

	// Convert public key to address
	address := crypto.PubkeyToAddress(*pubKey)
	return address, nil
}

func (api *APIImpl) ParseABI(txData []byte, chainId *big.Int) {
	//txData := txn.GetData()
	//hexData := hexutil.Bytes(txData)
	//hexUtilityData := (*hexutility.Bytes)(&hexData)
	if len(txData) < 4 {
		return
	}

	//signer := types.LatestSignerForChainID(chainId)
	//fromAddress, _ := txn.Sender(*signer)
	//log.Info(fmt.Sprintf("hash:%v, sender:%v, to:%v, data:%v", txn.Hash(), fromAddress, txn.GetTo(), hexUtilityData))

	parsedABI, err := abi.JSON(bytes.NewReader([]byte(PayABIData.ABI)))
	if err != nil {
		log.Error("zjg, Failed to parse ABI", "error", err)
		return
	}

	method, err := parsedABI.MethodById(txData[:4])
	if err != nil {
		log.Debug(fmt.Sprintf("parsedABI failed %v", err))
		return
	}

	if method.Name != TargetMethod {
		return
	}
	log.Info(fmt.Sprintf("zjg, method name: %s, id:%#x, txdata:%v", method.Name, string(method.ID), hexutil.Encode(txData)))
	args := make(map[string]interface{})
	err = method.Inputs.UnpackIntoMap(args, txData[4:])
	if err != nil {
		log.Error("zjg, UnpackIntoMap failed %v", err)
		return
	}
	rawOps := args["ops"].([]struct {
		Sender             libcommon.Address "json:\"sender\""
		Nonce              *big.Int          "json:\"nonce\""
		InitCode           []uint8           "json:\"initCode\""
		CallData           []uint8           "json:\"callData\""
		AccountGasLimits   [32]uint8         "json:\"accountGasLimits\""
		PreVerificationGas *big.Int          "json:\"preVerificationGas\""
		GasFees            [32]uint8         "json:\"gasFees\""
		PaymasterAndData   []uint8           "json:\"paymasterAndData\""
		Signature          []uint8           "json:\"signature\""
	})

	var userOps []PackedUserOperation
	for _, op := range rawOps {
		userOps = append(userOps, PackedUserOperation{
			Sender:             op.Sender,
			Nonce:              op.Nonce,
			InitCode:           op.InitCode,
			CallData:           op.CallData,
			AccountGasLimits:   op.AccountGasLimits,
			PreVerificationGas: op.PreVerificationGas,
			GasFees:            op.GasFees,
			PaymasterAndData:   op.PaymasterAndData,
			Signature:          op.Signature,
		})
	}
	for _, op := range userOps {
		_, signature, validationData, err := api.decodeSignature(op.Signature)
		//signatureHash := "keccak256(abi.encode(userOp.hash(), address(this), block.chainid));"
		userOpHash, _ := op.Hash()
		entryPointAddress := libcommon.HexToAddress("0x0FA5445ed0a18817f714D33d89F44eE4452aEa77")
		chainID := big.NewInt(195)

		Uint256, _ := abi.NewType("uint256", "", nil)
		Address, _ := abi.NewType("address", "", nil)
		Bytes, _ := abi.NewType("bytes", "", nil)
		//Byte32, _ := abi.NewType("bytes32", "", nil)
		// Encode the full data for signature hash
		signatureArguments := abi.Arguments{
			{Type: Bytes},   // userOp.hash() (Keccak256 is 32 bytes)
			{Type: Address}, // address(this)
			{Type: Uint256}, // block.chainid
		}

		signatureData, err := signatureArguments.Pack(userOpHash, entryPointAddress, chainID)
		if err != nil {
			log.Error("zjg, Failed to encode signature data:", err)
			return
		}

		// Hash the encoded data
		signatureHashReslut := sha3.NewLegacyKeccak256()
		signatureHashReslut.Write(signatureData)
		signatureHash := signatureHashReslut.Sum(nil)
		log.Info(fmt.Sprintf("zjg, Signature Hash: 0x%x\n", signatureHash))

		// totalHash := "keccak256(abi.encode(userOpHash, validationData)"
		arguments := abi.Arguments{
			{Type: Bytes},
			{Type: Uint256},
		}
		log.Info(fmt.Sprintf("zjg, userOpHash: 0x%x, validationData:%v", userOpHash, validationData))
		encodedData, err := arguments.Pack(userOpHash, validationData)
		if err != nil {
			log.Error("zjg, Failed to encode data:", err)
			return
		}

		totalHash := crypto.Keccak256(encodedData)

		prefix := []byte("\x19Ethereum Signed Message:\n32")
		data := append(prefix, totalHash...)

		hash := sha3.NewLegacyKeccak256()
		hash.Write(data)
		signedHash := hash.Sum(nil)
		log.Info(fmt.Sprintf("zjg, Signed Hash: 0x%x\n", signedHash))

		//userSignature, eoaSignature := abi.decode(signature)
		// Define ABI argument types
		arguments = abi.Arguments{
			{Type: Bytes}, // userSignature
			{Type: Bytes}, // eoaSignature
		}

		// Decode the signature
		var decodedData []interface{}
		decodedData, _ = arguments.Unpack(signature)
		if err != nil {
			log.Error("zjg, Failed to decode signature: %v", err)
			return
		}

		// Extract userSignature and eoaSignature
		for i, v := range decodedData {
			log.Info(fmt.Sprintf("zjg, decodedData[%d]: %v", i, v))
		}

		log.Info(fmt.Sprintf("decodedData[0] 类型: %T\n", decodedData[0]))
		userSignature, ok := decodedData[0].([]byte)
		if !ok {
			log.Error(fmt.Sprintf("zjg, decodedData[0] []byte"))
			return
		}
		eoaSignature, ok := decodedData[1].([]byte)
		if !ok {
			log.Error(fmt.Sprintf("zjg, decodedData[1] []byte"))
			return
		}
		if len(userSignature) == 0 || len(eoaSignature) == 0 {
			//log.Error(fmt.Sprintf("zjg, userSignature or eoaSignature is empty,%v,%v"), len(userSignature), len(eoaSignature))
			log.Error("zjg, User Signature: 0x%x\n", userSignature)
			log.Error("zjg, EOA Signature: 0x%x\n", eoaSignature)
			return
		}

		//1 recover
		//ECDSA.recover(signedHash, eoaSignature)
		//add, err := recoverSigner(signedHash, eoaSignature)
		{
			var testAddrHex = "970e8128ab834e8eac17ab8e3812f010678cf791"
			var testPrivHex = "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032"
			key, _ := crypto.HexToECDSA(testPrivHex)
			addr := libcommon.HexToAddress(testAddrHex)

			if err != nil {
				log.Info(fmt.Sprintf("zjg, Failed to recover signer: %v", err))
			}

			msg := crypto.Keccak256([]byte("foo"))
			sig, err := crypto.Sign(msg, key)
			if err != nil {
				log.Error("Sign error: %s", err)
			}
			log.Info(fmt.Sprintf("zjg, msg len:%v, sig len:%v", len(msg), len(sig)))
			log.Info(fmt.Sprintf("zjg, msg: 0x%x", msg))
			log.Info(fmt.Sprintf("zjg, Signature: 0x%x", sig))
			recoveredPub, err := crypto.Ecrecover(msg, sig)
			pubKey, _ := crypto.UnmarshalPubkeyStd(recoveredPub)
			recoveredAddr := crypto.PubkeyToAddress(*pubKey)
			if addr != recoveredAddr {
				log.Error("Address mismatch: want: %x have: %x", addr, recoveredAddr)
			}
			log.Info(fmt.Sprintf("zjg, Recovered signer: %v\n", recoveredAddr))
		}
		log.Info(fmt.Sprintf("zjg, signedHash len:%v, eoaSignature len:%v", len(signedHash), len(eoaSignature)))
		log.Info(fmt.Sprintf("zjg, signedHash: 0x%x", signedHash))
		log.Info(fmt.Sprintf("zjg, EOA Signature: 0x%x", eoaSignature))
		pub, err := crypto.Ecrecover(signedHash, eoaSignature)
		if err != nil {
			log.Error("zjg, Failed to recover signer: %v", err)
		} else {
			log.Info("zjg, Recovered signer: %v\n", pub)
		}
	}
}

func (api *APIImpl) handleSignature(signature []uint8) {

}

type Result struct {
	pubKey         []byte
	sigData        []byte
	validationData *big.Int
}

func (api *APIImpl) decodeSignature(sig []byte) ([]byte, []byte, *big.Int, error) {
	log.Info(fmt.Sprintf("zjg, signature len:%v", len(sig)))
	parsedABI, err := abi.JSON(bytes.NewReader([]byte(HelpABIData.ABI)))
	if err != nil {
		log.Error("zjg, Failed to parse ABI", "error", err)
		return nil, nil, nil, err
	}

	outputs, err := parsedABI.Unpack(TargetMethod2, sig)
	if err != nil {
		log.Error("zjg, UnpackIntoInterface failed", "error", err)
		return nil, nil, nil, fmt.Errorf("failed to decode signature: %v", err)
	}

	result := Result{
		pubKey:         outputs[0].([]byte),
		sigData:        outputs[1].([]byte),
		validationData: outputs[2].(*big.Int),
	}
	log.Info(fmt.Sprintf("zjg, decoded signature: %x, %x, %v", result.pubKey, result.sigData, result.validationData))

	return result.pubKey, result.sigData, result.validationData, nil
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
