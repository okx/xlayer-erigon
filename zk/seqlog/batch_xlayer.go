package seqlog

import (
	"fmt"
	"sync"
	"time"
)

var batchLogger *batchLogInstance
var batchOnce sync.Once

type batchLogInstance struct {
	BatchNum      uint64
	BlockCount    uint64
	TxCount       uint64
	ClosingReason string
	TotalDuration time.Duration
	StepLog       string
}

func GetBatchLogger() *batchLogInstance {
	batchOnce.Do(func() {
		batchLogger = &batchLogInstance{}
		batchLogger.init()
	})
	return batchLogger
}

func (b *batchLogInstance) init() {
	//BlockNum > 0 means there is a wip log inside blockLogger
	//BlockNum = 0 means the blockLogger is empty
	b.BatchNum = 0
	b.BlockCount = 0
	b.TxCount = 0
	b.ClosingReason = ""
	b.TotalDuration = 0
	b.StepLog = ""
}

func (b *batchLogInstance) SetBlockNum(batchNum uint64) {
	b.BatchNum = batchNum
}

func (b *batchLogInstance) AccmuTxCount(txCount uint64) {
	b.TxCount += txCount
}

func (b *batchLogInstance) AccmuBlockCount() {
	b.BlockCount += 1
}

func (b *batchLogInstance) SetClosingReason(closingReason string) {
	b.ClosingReason = closingReason
}

func (b *batchLogInstance) SetTotalDuration(totalDuration time.Duration) {
	b.TotalDuration = totalDuration
}

func (b *batchLogInstance) AppendBlockLog(blockNum uint64, stepDuration time.Duration) {
	stepFloatDuration := float64(stepDuration.Microseconds()) / 1000
	b.StepLog = b.StepLog + "," + fmt.Sprintf("Block %d<%.2fms>", blockNum, stepFloatDuration)
}

func (b *batchLogInstance) AppendCommitLog(stepDuration time.Duration) {
	stepFloatDuration := float64(stepDuration.Microseconds()) / 1000
	b.StepLog = b.StepLog + "," + fmt.Sprintf("BatchCommit<%.2fms>", stepFloatDuration)
}

func (b *batchLogInstance) PrintLogAndFlush() string {
	totalFloatDuration := float64(b.TotalDuration.Microseconds()) / 1000.0
	itemLog := fmt.Sprintf("[Batch Log] Batch<%d>,ClosingReason<%s>,Block<%d>,Tx<%d>,TotalDuration<%.2fms>", b.BatchNum, b.ClosingReason, b.BlockCount, b.TxCount, totalFloatDuration)
	overallLog := itemLog + b.StepLog
	//Flush blockLogger
	b.init()
	return overallLog
}
