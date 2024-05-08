package vm

import "math/big"

// InnerTxBasic stores the basic field of an inner tx.
// NOTE: DON'T change this struct for:
// 1. It will be written to database, and must be keep the same type When reading history data from db
// 2. It will be returned by rpc method
type InnerTxBasic struct {
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

type BlockInnerData struct {
	BlockHash string
	TxHashes  []string
	TxMap     map[string][]*InnerTxBasic
}

// InnerTx store all field of an inner tx, you can change/add those field as you will.
type InnerTx struct {
	InnerTxBasic

	//FromNonce       uint64
	//Create2Salt     string
	//Create2CodeHash string
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
