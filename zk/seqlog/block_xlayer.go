package seqlog

import (
	"fmt"
	"sync"
	"time"
)

var blockLogger *blockLogInstance
var blockOnce sync.Once

type blockLogInstance struct {
	BlockNum      uint64
	TxCount       uint64
	TotalDuration time.Duration
	StepLog       string
}

func GetBlockLogger() *blockLogInstance {
	blockOnce.Do(func() {
		blockLogger = &blockLogInstance{}
		blockLogger.init()
	})
	return blockLogger
}

func (b *blockLogInstance) init() {
	b.BlockNum = 0
	b.TxCount = 0
	b.TotalDuration = 0
	b.StepLog = ""
}

func (b *blockLogInstance) SetBlockNum(blockNum uint64) {

	b.BlockNum = blockNum
}

func (b *blockLogInstance) SetTxCount(txCount uint64) {
	b.TxCount = txCount
}

func (b *blockLogInstance) SetTotalDuration(totalDuration time.Duration) {
	b.TotalDuration = totalDuration
}

func (b *blockLogInstance) AppendStepLog(stepTag string, stepDuration time.Duration) {
	stepFloatDuration := float64(stepDuration.Microseconds()) / 1000
	b.StepLog = b.StepLog + "," + fmt.Sprintf("%s<%.2fms>", stepTag, stepFloatDuration)
}

func (b *blockLogInstance) PrintLogAndFlush() string {
	totalFloatDuration := float64(b.TotalDuration.Microseconds()) / 1000.0
	itemLog := fmt.Sprintf("[Block Log] Block<%d>,Tx<%d>,TotalDuration<%.2fms>", b.BlockNum, b.TxCount, totalFloatDuration)
	overallLog := itemLog + b.StepLog
	b.init()
	return overallLog
}

// stepTag Name
const (
	AddTxs           = "attemptAddTransaction"
	WaitBlockTimeOut = "waitingBlockTimeout"
	WaitBatchTimeOut = "waitingBatchTimeout"
	PbState          = "postBlockStateHandling"
	ZkInc            = "zkIncrementIntermediateHashes"
	DoFin            = "finaliseBlockWrite"
	Save2DB          = "commitToDB"
)
