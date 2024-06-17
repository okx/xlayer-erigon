package vm

import (
	libcommon "github.com/gateway-fm/cdk-erigon-lib/common"
	"github.com/gateway-fm/cdk-erigon-lib/common/hexutility"
	zktypes "github.com/ledgerwatch/erigon/zk/types"
	"math/big"
	"strconv"
)

const (
	CALL_TYP         = "call"
	CALLCODE_TYP     = "callcode"
	DELEGATECALL_TYP = "delegatecall"
	STATICCAL_TYP    = "staticcall"
	CREATE_TYP       = "create"
	CREATE2_TYP      = "create2"
	SUICIDE_TYP      = "suicide"
)

//// InnerTx stores the basic field of an inner tx.
//// NOTE: DON'T change this struct for:
//// 1. It will be written to database, and must be keep the same type When reading history data from db
//// 2. It will be returned by rpc method
//type InnerTx struct {
//	Dept          big.Int `json:"dept"`
//	InternalIndex big.Int `json:"internal_index"`
//	CallType      string  `json:"call_type"`
//	Name          string  `json:"name"`
//	TraceAddress  string  `json:"trace_address"`
//	CodeAddress   string  `json:"code_address"`
//	From          string  `json:"from"`
//	To            string  `json:"to"`
//	Input         string  `json:"input"`
//	Output        string  `json:"output"`
//	IsError       bool    `json:"is_error"`
//	Gas           uint64  `json:"gas"`
//	GasUsed       uint64  `json:"gas_used"`
//	Value         string  `json:"value"`
//	ValueWei      string  `json:"value_wei"`
//	Error         string  `json:"error"`
//}

type InnerTxMeta struct {
	index     int
	lastDepth int
	indexMap  map[int]int
	InnerTxs  []*zktypes.InnerTx
}

func (evm *EVM) GetInnerTxMeta() *InnerTxMeta {
	return evm.innerTxMeta
}

func (evm *EVM) AddInnerTx(innerTx *zktypes.InnerTx) {
	evm.innerTxMeta.InnerTxs = append(evm.innerTxMeta.InnerTxs, innerTx)
}

func beforeOp(
	interpreter *EVMInterpreter,
	callTyp string,
	fromAddr libcommon.Address,
	toAddr *libcommon.Address,
	codeAddr *libcommon.Address,
	input []byte,
	gas uint64,
	value *big.Int) (*zktypes.InnerTx, int) {
	innerTx := &zktypes.InnerTx{
		CallType: callTyp,
		From:     fromAddr.String(),
		ValueWei: value.String(),
		Gas:      gas,
		IsError:  false,
	}

	if toAddr != nil {
		innerTx.To = toAddr.String()
	}
	if codeAddr != nil {
		innerTx.CodeAddress = codeAddr.String()
	}

	if input != nil {
		innerTx.Input = hexutility.Encode(input)
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

	newIndex := len(interpreter.evm.GetInnerTxMeta().InnerTxs) - 1
	if newIndex < 0 {
		newIndex = 0
	}

	return innerTx, newIndex
}

func afterOp(interpreter *EVMInterpreter, opType string, gas_used uint64, newIndex int, innerTx *zktypes.InnerTx, addr *libcommon.Address, err error) {
	innerTx.GasUsed = gas_used
	if err != nil {
		innerTxMeta := interpreter.evm.GetInnerTxMeta()
		for _, innerTx := range innerTxMeta.InnerTxs[newIndex:] {
			innerTx.IsError = true
		}
	}

	switch opType {
	case "create":
		innerTx.To = addr.String()
	}
}
