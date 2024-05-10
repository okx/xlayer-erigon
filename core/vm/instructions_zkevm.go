package vm

import (
	"math/big"
	"strconv"

	libcommon "github.com/gateway-fm/cdk-erigon-lib/common"
	"github.com/holiman/uint256"
	"github.com/ledgerwatch/erigon/common"
	"github.com/ledgerwatch/erigon/core/types"
	"github.com/ledgerwatch/erigon/params"
)

func opCallDataLoad_zkevmIncompatible(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	x := scope.Stack.Peek()
	if offset, overflow := x.Uint64WithOverflow(); !overflow {
		data := getData(scope.Contract.Input, offset, 32)
		if len(scope.Contract.Input) == 0 {
			data = getData(scope.Contract.Code, offset, 32)
		}
		x.SetBytes(data)
	} else {
		x.Clear()
	}
	return nil, nil
}

func opCallDataCopy_zkevmIncompatible(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	var (
		memOffset  = scope.Stack.Pop()
		dataOffset = scope.Stack.Pop()
		length     = scope.Stack.Pop()
	)
	dataOffset64, overflow := dataOffset.Uint64WithOverflow()
	if overflow {
		dataOffset64 = 0xffffffffffffffff
	}
	// These values are checked for overflow during gas cost calculation
	memOffset64 := memOffset.Uint64()
	length64 := length.Uint64()

	if len(scope.Contract.Input) == 0 {
		scope.Memory.Set(memOffset64, length64, getData(scope.Contract.Code, dataOffset64, length64))
	} else {
		scope.Memory.Set(memOffset64, length64, getData(scope.Contract.Input, dataOffset64, length64))
	}

	return nil, nil
}

func opExtCodeHash_zkevm(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	slot := scope.Stack.Peek()
	address := libcommon.Address(slot.Bytes20())
	ibs := interpreter.evm.IntraBlockState()
	if ibs.GetCodeSize(address) == 0 {
		slot.SetBytes(libcommon.Hash{}.Bytes())
	} else {
		slot.SetBytes(ibs.GetCodeHash(address).Bytes())
	}
	return nil, nil
}

func opBlockhash_zkevm(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	num := scope.Stack.Peek()
	num64, overflow := num.Uint64WithOverflow()
	if overflow {
		num.Clear()
		return nil, nil
	}

	ibs := interpreter.evm.IntraBlockState()
	hash := ibs.GetBlockStateRoot(num64)

	num.SetFromBig(hash.Big())

	return nil, nil
}

func opNumber_zkevm(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	ibs := interpreter.evm.IntraBlockState()
	num := ibs.GetBlockNumber()
	scope.Stack.Push(num)
	return nil, nil
}

func opDifficulty_zkevm(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	zeroInt := new(big.Int).SetUint64(0)
	v, _ := uint256.FromBig(zeroInt)
	scope.Stack.Push(v)
	return nil, nil
}

func opStaticCall_zkevm(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	// Pop gas. The actual gas is in interpreter.evm.callGasTemp.
	stack := scope.Stack
	// We use it as a temporary value
	temp := stack.Pop()
	gas := interpreter.evm.CallGasTemp()
	// Pop other call parameters.
	addr, inOffset, inSize, retOffset, retSize := stack.Pop(), stack.Pop(), stack.Pop(), stack.Pop(), stack.Pop()
	toAddr := libcommon.Address(addr.Bytes20())
	// Get arguments from the memory.
	args := scope.Memory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))
	innerTx, newIndex := beforeOp(interpreter, STATICCALL.String(), scope.Contract.Address(), &toAddr, nil, gas, big.NewInt(0))
	ret, returnGas, err := interpreter.evm.StaticCall(scope.Contract, toAddr, args, gas)
	afterOp(interpreter, "call", newIndex, innerTx, nil, err)
	if err != nil {
		temp.Clear()
	} else {
		temp.SetOne()
	}
	stack.Push(&temp)
	if err == nil || IsErrTypeRevert(err) {
		ret = common.CopyBytes(ret)
		scope.Memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}

	scope.Contract.Gas += returnGas

	//[zkevm] do not overryde returnData if reverted
	if !IsErrTypeRevert(err) {
		interpreter.returnData = ret
	}

	return ret, nil
}

// removed the actual self destruct at the end
func opSendAll_zkevm(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	if interpreter.readOnly {
		return nil, ErrWriteProtection
	}
	beneficiary := scope.Stack.Pop()
	callerAddr := scope.Contract.Address()
	beneficiaryAddr := libcommon.Address(beneficiary.Bytes20())
	balance := interpreter.evm.IntraBlockState().GetBalance(callerAddr)
	if interpreter.evm.Config().Debug {
		if interpreter.cfg.Debug {
			interpreter.cfg.Tracer.CaptureEnter(SELFDESTRUCT, callerAddr, beneficiaryAddr, false /* precompile */, false /* create */, []byte{}, 0, balance, nil /* code */)
			interpreter.cfg.Tracer.CaptureExit([]byte{}, 0, nil)
		}
	}

	if beneficiaryAddr != callerAddr {
		innerTx, newIndex := beforeOp(interpreter, SENDALL.String(), scope.Contract.Address(), &beneficiaryAddr, nil, 0, balance.ToBig())
		interpreter.evm.IntraBlockState().AddBalance(beneficiaryAddr, balance)
		interpreter.evm.IntraBlockState().SubBalance(callerAddr, balance)
		afterOp(interpreter, "send_all", newIndex, innerTx, nil, nil)
	}
	return nil, errStopToken
}

// [zkEvm] log data length must be a multiple of 32, if not - fill 0 at the end until it is
func makeLog_zkevm(size int) executionFunc {
	return func(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
		if interpreter.readOnly {
			return nil, ErrWriteProtection
		}
		topics := make([]libcommon.Hash, size)
		stack := scope.Stack
		mStart, mSize := stack.Pop(), stack.Pop()
		for i := 0; i < size; i++ {
			addr := stack.Pop()
			topics[i] = addr.Bytes32()
		}

		d := scope.Memory.GetCopy(int64(mStart.Uint64()), int64(mSize.Uint64()))

		// [zkEvm] fill 0 at the end
		dataLen := len(d)
		lenMod32 := dataLen & 31
		if lenMod32 != 0 {
			d = append(d, make([]byte, 32-lenMod32)...)
		}

		interpreter.evm.IntraBlockState().AddLog_zkEvm(&types.Log{
			Address: scope.Contract.Address(),
			Topics:  topics,
			Data:    d,
			// This is a non-consensus field, but assigned here because
			// core/state doesn't know the current block number.
			BlockNumber: interpreter.evm.Context().BlockNumber,
		})

		return nil, nil
	}
}

func opCreate_zkevm(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	if interpreter.readOnly {
		return nil, ErrWriteProtection
	}
	var (
		value  = scope.Stack.Pop()
		offset = scope.Stack.Pop()
		size   = scope.Stack.Peek()
		input  = scope.Memory.GetCopy(int64(offset.Uint64()), int64(size.Uint64()))
		gas    = scope.Contract.Gas
	)
	if interpreter.evm.ChainRules().IsTangerineWhistle {
		gas -= gas / 64
	}
	// reuse size int for stackvalue
	stackvalue := size

	scope.Contract.UseGas(gas)

	innerTx, newIndex := beforeOp(interpreter, CREATE.String(), scope.Contract.Address(), nil, nil, gas, value.ToBig())
	res, addr, returnGas, suberr := interpreter.evm.Create(scope.Contract, input, gas, &value, 0)
	afterOp(interpreter, "create", newIndex, innerTx, &addr, suberr)

	// Push item on the stack based on the returned error. If the ruleset is
	// homestead we must check for CodeStoreOutOfGasError (homestead only
	// rule) and treat as an error, if the ruleset is frontier we must
	// ignore this error and pretend the operation was successful.
	if interpreter.evm.ChainRules().IsHomestead && suberr == ErrCodeStoreOutOfGas {
		stackvalue.Clear()
	} else if suberr != nil && suberr != ErrCodeStoreOutOfGas {
		stackvalue.Clear()
	} else {
		stackvalue.SetBytes(addr.Bytes())
	}
	scope.Contract.Gas += returnGas

	if IsErrTypeRevert(suberr) {
		interpreter.returnData = res // set REVERT data to return data buffer
		return res, nil
	}
	interpreter.returnData = nil // clear dirty return data buffer
	return nil, nil
}

func opCreate2_zkevm(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	if interpreter.readOnly {
		return nil, ErrWriteProtection
	}
	var (
		endowment    = scope.Stack.Pop()
		offset, size = scope.Stack.Pop(), scope.Stack.Pop()
		salt         = scope.Stack.Pop()
		input        = scope.Memory.GetCopy(int64(offset.Uint64()), int64(size.Uint64()))
		gas          = scope.Contract.Gas
	)

	// Apply EIP150
	gas -= gas / 64
	scope.Contract.UseGas(gas)
	// reuse size int for stackvalue
	stackValue := size
	innerTx, newIndex := beforeOp(interpreter, CREATE2.String(), scope.Contract.Address(), nil, nil, gas, endowment.ToBig())
	res, addr, returnGas, suberr := interpreter.evm.Create2(scope.Contract, input, gas, &endowment, &salt)
	afterOp(interpreter, "create", newIndex, innerTx, &addr, suberr)

	// Push item on the stack based on the returned error.
	if suberr != nil {
		stackValue.Clear()
	} else {
		stackValue.SetBytes(addr.Bytes())
	}

	scope.Stack.Push(&stackValue)
	scope.Contract.Gas += returnGas

	if IsErrTypeRevert(suberr) {
		interpreter.returnData = res // set REVERT data to return data buffer
		return res, nil
	}
	interpreter.returnData = nil // clear dirty return data buffer
	return nil, nil
}

func opCall_zkevm(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	stack := scope.Stack
	// Pop gas. The actual gas in interpreter.evm.callGasTemp.
	// We can use this as a temporary value
	temp := stack.Pop()
	gas := interpreter.evm.CallGasTemp()
	// Pop other call parameters.
	addr, value, inOffset, inSize, retOffset, retSize := stack.Pop(), stack.Pop(), stack.Pop(), stack.Pop(), stack.Pop(), stack.Pop()
	toAddr := libcommon.Address(addr.Bytes20())
	// Get the arguments from the memory.
	args := scope.Memory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))

	if !value.IsZero() {
		if interpreter.readOnly {
			return nil, ErrWriteProtection
		}
		gas += params.CallStipend
	}

	innerTx, newIndex := beforeOp(interpreter, CALL.String(), scope.Contract.Address(), &toAddr, nil, gas, value.ToBig())
	ret, returnGas, err := interpreter.evm.Call(scope.Contract, toAddr, args, gas, &value, false /* bailout */, 0)
	afterOp(interpreter, "call", newIndex, innerTx, nil, err)

	if err != nil {
		temp.Clear()
	} else {
		temp.SetOne()
	}
	stack.Push(&temp)
	if err == nil || IsErrTypeRevert(err) {
		ret = common.CopyBytes(ret)
		scope.Memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}

	scope.Contract.Gas += returnGas

	interpreter.returnData = ret
	return ret, nil
}

func opCallCode_zkevm(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	// Pop gas. The actual gas is in interpreter.evm.callGasTemp.
	stack := scope.Stack
	// We use it as a temporary value
	temp := stack.Pop()
	gas := interpreter.evm.CallGasTemp()
	// Pop other call parameters.
	addr, value, inOffset, inSize, retOffset, retSize := stack.Pop(), stack.Pop(), stack.Pop(), stack.Pop(), stack.Pop(), stack.Pop()
	toAddr := libcommon.Address(addr.Bytes20())
	// Get arguments from the memory.
	args := scope.Memory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))

	if !value.IsZero() {
		gas += params.CallStipend
	}

	innerTx, newIndex := beforeOp(interpreter, CALLCODE.String(), scope.Contract.Address(), &toAddr, &toAddr, gas, value.ToBig())
	ret, returnGas, err := interpreter.evm.CallCode(scope.Contract, toAddr, args, gas, &value)
	afterOp(interpreter, "call", newIndex, innerTx, nil, err)
	if err != nil {
		temp.Clear()
	} else {
		temp.SetOne()
	}
	stack.Push(&temp)
	if err == nil || IsErrTypeRevert(err) {
		ret = common.CopyBytes(ret)
		scope.Memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}

	scope.Contract.Gas += returnGas

	interpreter.returnData = ret
	return ret, nil
}

func opDelegateCall_zkevm(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	stack := scope.Stack
	// Pop gas. The actual gas is in interpreter.evm.callGasTemp.
	// We use it as a temporary value
	temp := stack.Pop()
	gas := interpreter.evm.CallGasTemp()
	// Pop other call parameters.
	addr, inOffset, inSize, retOffset, retSize := stack.Pop(), stack.Pop(), stack.Pop(), stack.Pop(), stack.Pop()
	toAddr := libcommon.Address(addr.Bytes20())
	// Get arguments from the memory.
	args := scope.Memory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))

	innerTx, newIndex := beforeOp(interpreter, DELEGATECALL.String(), scope.Contract.Address(), &toAddr, nil, gas, big.NewInt(0))
	ret, returnGas, err := interpreter.evm.DelegateCall(scope.Contract, toAddr, args, gas)

	innerTx.TraceAddress = scope.Contract.CallerAddress.Hash().String()
	innerTx.ValueWei = scope.Contract.value.String()
	afterOp(interpreter, "call", newIndex, innerTx, nil, err)

	if err != nil {
		temp.Clear()
	} else {
		temp.SetOne()
	}
	stack.Push(&temp)
	if err == nil || IsErrTypeRevert(err) {
		ret = common.CopyBytes(ret)
		scope.Memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}

	scope.Contract.Gas += returnGas

	interpreter.returnData = ret
	return ret, nil
}

func beforeOp(interpreter *EVMInterpreter, name string, fromAddr libcommon.Address, toAddr *libcommon.Address, codeAddr *libcommon.Address, gas uint64, value *big.Int) (*InnerTx, int) {
	innerTx := &InnerTx{
		CallType: name,
		From:     fromAddr.Hash().String(),
		ValueWei: value.String(),
		GasUsed:  gas,
		IsError:  false,
	}

	if toAddr != nil {
		innerTx.To = toAddr.Hash().String()
	}
	if codeAddr != nil {
		innerTx.CodeAddress = codeAddr.Hash().String()
	}

	innerTxMeta := interpreter.evm.GetInnerTxMeta()
	if innerTxMeta == nil {
		// TODO
	}
	if interpreter.Depth() == innerTxMeta.lastDepth {
		innerTxMeta.index++
		innerTxMeta.indexMap[interpreter.Depth()] = innerTxMeta.index
	} else if interpreter.Depth() < innerTxMeta.lastDepth {
		innerTxMeta.index = innerTxMeta.indexMap[interpreter.Depth()] + 1
		innerTxMeta.indexMap[interpreter.Depth()] = innerTxMeta.index
		innerTxMeta.lastDepth = interpreter.Depth()
	} else if interpreter.Depth() > innerTxMeta.lastDepth {
		innerTxMeta.index = 0
		innerTxMeta.indexMap[interpreter.Depth()] = 0
		innerTxMeta.lastDepth = interpreter.Depth()
	}
	for i := 1; i <= innerTxMeta.lastDepth; i++ {
		innerTx.Name = innerTx.Name + "_" + strconv.Itoa(innerTxMeta.indexMap[i])
	}
	innerTx.Name = innerTx.CallType + innerTx.Name
	innerTx.Dept = *big.NewInt(int64(interpreter.Depth()))
	innerTx.InternalIndex = *big.NewInt(int64(innerTxMeta.index))

	interpreter.evm.AddInnerTx(innerTx)

	newIndex := len(innerTxMeta.InnerTxs) - 1
	if newIndex < 0 {
		newIndex = 0
	}

	// TODO
	//interpreter.evm.SetInnerTxMeta(innerTxMeta)
	return innerTx, newIndex
}

func afterOp(interpreter *EVMInterpreter, opType string, newIndex int, innerTx *InnerTx, addr *libcommon.Address, err error) {
	if err != nil {
		innerTxMeta := interpreter.evm.GetInnerTxMeta()
		for _, innerTx := range innerTxMeta.InnerTxs[newIndex:] {
			innerTx.IsError = true
		}
	}

	switch opType {
	case "create":
		innerTx.To = addr.Hash().String()
	}
}
