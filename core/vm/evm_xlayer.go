package vm

import "math/big"

// InnerTx stores the basic field of an inner tx.
// NOTE: DON'T change this struct for:
// 1. It will be written to database, and must be keep the same type When reading history data from db
// 2. It will be returned by rpc method
type InnerTx struct {
	Dept          big.Int `json:"dept"`
	InternalIndex big.Int `json:"internal_index"`
	CallType      string  `json:"call_type"`
	Name          string  `json:"name"`
	TraceAddress  string  `json:"trace_address"`
	CodeAddress   string  `json:"code_address"`
	From          string  `json:"from"`
	To            string  `json:"to"`
	Input         string  `json:"input"`
	Output        string  `json:"output"`
	IsError       bool    `json:"is_error"`
	GasUsed       uint64  `json:"gas_used"`
	Value         string  `json:"value"`
	ValueWei      string  `json:"value_wei"`
	Error         string  `json:"error"`
}

type InnerTxMeta struct {
	index     int
	lastDepth int
	indexMap  map[int]int
	InnerTxs  []*InnerTx
}

func (evm *EVM) GetInnerTxMeta() *InnerTxMeta {
	return evm.innerTxMeta
}

func (evm *EVM) AddInnerTx(innerTx *InnerTx) {
	evm.innerTxMeta.InnerTxs = append(evm.innerTxMeta.InnerTxs, innerTx)
}
