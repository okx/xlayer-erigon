package rocksdb

import (
	"context"
	"fmt"
	"github.com/ledgerwatch/erigon-lib/common/dbg"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/log/v3"
	"github.com/linxGnu/grocksdb"
	"golang.org/x/sync/semaphore"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

var pathDbMap = map[string]kv.RoDB{}
var pathDbMapLock sync.Mutex

type RocksKV struct {
	db  *grocksdb.DB
	log log.Logger

	readTxLimiter  *semaphore.Weighted // does limit amount of concurrent Ro transactions - in most casess runtime.NumCPU() is good value for this channel capacity - this channel can be shared with other components (like Decompressor)
	writeTxLimiter *semaphore.Weighted
	opts           RocksDBOpts
	txSize         uint64
	closed         atomic.Bool
	path           string

	txsCount              uint
	txsCountMutex         *sync.Mutex
	txsAllDoneOnCloseCond *sync.Cond

	leakDetector *dbg.LeakDetector

	// MaxBatchSize is the maximum size of a batch. Default value is
	// copied from DefaultMaxBatchSize in Open.
	//
	// If <=0, disables batching.
	//
	// Do not change concurrently with calls to Batch.
	MaxBatchSize int

	// MaxBatchDelay is the maximum delay before a batch starts.
	// Default value is copied from DefaultMaxBatchDelay in Open.
	//
	// If <=0, effectively disables batching.
	//
	// Do not change concurrently with calls to Batch.
	MaxBatchDelay time.Duration

	batchMu sync.Mutex

	cfHandles map[string]*grocksdb.ColumnFamilyHandle
}

func (kv *RocksKV) Close() {
	if ok := kv.closed.CompareAndSwap(false, true); !ok {
		return
	}
	kv.db.Close()

	if kv.opts.inMem {
		if err := os.RemoveAll(kv.opts.path); err != nil {
			kv.log.Warn("failed to remove in-mem db file", "err", err)
		}
	}

	removeFromPathDbMap(kv.path)
}

func (kv *RocksKV) ReadOnly() bool {
	return kv.opts.readOnly
}

func (kv *RocksKV) View(ctx context.Context, f func(tx kv.Tx) error) error {
	//TODO implement me
	panic("implement me")
}

func (kv *RocksKV) BeginRo(ctx context.Context) (kv.Tx, error) {
	//TODO implement me
	panic("implement me")
}

func (kv *RocksKV) AllTables() kv.TableCfg {
	//TODO implement me
	panic("implement me")
}

func (kv *RocksKV) PageSize() uint64 {
	//TODO implement me
	panic("implement me")
}

func (kv *RocksKV) CHandle() unsafe.Pointer {
	//TODO implement me
	panic("implement me")
}

func (kv *RocksKV) Update(ctx context.Context, f func(tx kv.RwTx) error) error {
	tx, err := kv.BeginRw(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	err = f(tx)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (kv *RocksKV) UpdateNosync(ctx context.Context, f func(tx kv.RwTx) error) error {
	//TODO implement me
	panic("implement me")
}

func (kv *RocksKV) BeginRw(ctx context.Context) (kv.RwTx, error) {
	//TODO implement me
	panic("implement me")
}

func (kv *RocksKV) BeginRwNosync(ctx context.Context) (kv.RwTx, error) {
	//TODO implement me
	panic("implement me")
}

func removeFromPathDbMap(path string) {
	pathDbMapLock.Lock()
	defer pathDbMapLock.Unlock()
	delete(pathDbMap, path)
}

func addToPathDbMap(path string, db kv.RoDB) {
	pathDbMapLock.Lock()
	defer pathDbMapLock.Unlock()
	pathDbMap[path] = db
}

func (kv *RocksKV) beginRw(ctx context.Context, flags uint) (txn kv.RwTx, err error) {
	if kv.closed.Load() {
		return nil, fmt.Errorf("db is closed")
	}

	return &RocksTx{
		kv:  kv,
		ctx: ctx,
		id:  kv.leakDetector.Add(),
	}, nil
}

func (kv *RocksKV) beginRo(ctx context.Context) (txn kv.Tx, err error) {
	return nil, nil
}
