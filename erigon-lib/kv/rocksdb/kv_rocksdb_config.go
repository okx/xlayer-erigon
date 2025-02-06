package rocksdb

import (
	"context"
	"github.com/ledgerwatch/erigon-lib/common/dbg"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/log/v3"
	"github.com/linxGnu/grocksdb"
	"golang.org/x/sync/semaphore"
	"runtime"
	"sync"
)

type RocksDBOpts struct {
	log            log.Logger
	label          kv.Label
	verbosity      kv.DBVerbosityLvl
	readTxLimiter  *semaphore.Weighted
	writeTxLimiter *semaphore.Weighted
	inMem          bool
	readOnly       bool
	exclusive      bool
	path           string
}

func NewRocksDBOpts(log log.Logger) *RocksDBOpts {
	return &RocksDBOpts{
		log: log,
	}
}

func (opts RocksDBOpts) GetLabel() kv.Label { return opts.label }

func (opts RocksDBOpts) RoTxsLimiter(l *semaphore.Weighted) RocksDBOpts {
	opts.readTxLimiter = l
	return opts
}
func (opts RocksDBOpts) DBVerbosity(v kv.DBVerbosityLvl) RocksDBOpts {
	opts.verbosity = v
	return opts
}

func (opts RocksDBOpts) InMem() RocksDBOpts {
	opts.inMem = true
	return opts
}
func NewRocksDB(log log.Logger) RocksDBOpts {
	opts := RocksDBOpts{
		log:   log,
		label: kv.InMem,
	}
	return opts
}
func (opts RocksDBOpts) Readonly() RocksDBOpts {
	opts.readOnly = true
	return opts
}

func (opts RocksDBOpts) Exclusive() RocksDBOpts {
	opts.exclusive = true
	return opts
}

func (opts RocksDBOpts) Label(label kv.Label) RocksDBOpts {
	opts.label = label
	return opts
}

func (opts RocksDBOpts) Path(path string) RocksDBOpts {
	opts.path = path
	return opts
}
func (opts RocksDBOpts) Open(ctx context.Context) (kv.RwDB, error) {
	rocksDBOptions := grocksdb.NewDefaultOptions()
	rocksDBOptions.SetCreateIfMissing(true)
	rocksDBOptions.SetCreateIfMissingColumnFamilies(true)
	rocksDBOptions.SetInfoLogLevel(grocksdb.InfoLogLevel(opts.verbosity))

	rocksDB, err := grocksdb.OpenDb(rocksDBOptions, opts.path)
	if err != nil {
		return nil, err
	}

	if opts.readTxLimiter == nil {
		targetSemCount := int64(runtime.GOMAXPROCS(-1) * 16)
		opts.readTxLimiter = semaphore.NewWeighted(targetSemCount) // 1 less than max to allow unlocking to happen
	}

	if opts.writeTxLimiter == nil {
		targetSemCount := int64(runtime.GOMAXPROCS(-1)) - 1
		opts.writeTxLimiter = semaphore.NewWeighted(targetSemCount) // 1 less than max to allow unlocking to happen
	}
	txsCountMutex := &sync.Mutex{}

	kv := &RocksKV{
		opts:                  opts,
		log:                   opts.log,
		readTxLimiter:         opts.readTxLimiter,
		writeTxLimiter:        opts.writeTxLimiter,
		txsCountMutex:         txsCountMutex,
		txsAllDoneOnCloseCond: sync.NewCond(txsCountMutex),
		leakDetector:          dbg.NewLeakDetector("db."+opts.label.String(), dbg.SlowTx()),
		db:                    rocksDB,
	}

	addToPathDbMap(opts.path, kv)

	return kv, nil
}
