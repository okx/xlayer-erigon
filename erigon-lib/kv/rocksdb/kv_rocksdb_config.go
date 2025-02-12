package rocksdb

import (
	"context"
	"path/filepath"
	"sync"

	"github.com/ledgerwatch/erigon-lib/common/dbg"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/log/v3"
	"github.com/linxGnu/grocksdb"
)

type TableCfgFunc func(defaultCF kv.TableCfg) kv.TableCfg

func WithChaindataTables(defaultBuckets kv.TableCfg) kv.TableCfg {
	return defaultBuckets
}

type RocksDBOpts struct {
	blockSize      int64
	blockCache     int64
	statistics     bool
	maxOpenFiles   int64
	mmapRead       bool
	mmapWrite      bool
	unorderedWrite bool
	pipelinedWrite bool

	log       log.Logger
	label     kv.Label
	verbosity kv.DBVerbosityLvl
	path      string
	cfConfig  TableCfgFunc
}

func NewRocksDBOpts(log log.Logger) *RocksDBOpts {
	return &RocksDBOpts{
		log:      log,
		cfConfig: WithChaindataTables,
		label:    kv.InMem,
	}
}

func (opts RocksDBOpts) Open(ctx context.Context) (kv.RwDB, error) {
	rocksDBOptions := grocksdb.NewDefaultOptions()
	rocksDBOptions.SetCreateIfMissing(true)
	rocksDBOptions.SetCreateIfMissingColumnFamilies(true)
	rocksDBOptions.SetInfoLogLevel(grocksdb.InfoLogLevel(opts.verbosity))
	txsCountMutex := &sync.Mutex{}

	kvStore := &RocksKV{
		opts:                  opts,
		log:                   opts.log,
		txsCountMutex:         txsCountMutex,
		txsAllDoneOnCloseCond: sync.NewCond(txsCountMutex),
		leakDetector:          dbg.NewLeakDetector("db."+opts.label.String(), dbg.SlowTx()),
		cf:                    kv.TableCfg{},
	}

	customCf := opts.cfConfig(kv.ChaindataTablesCfg)
	for name, cfg := range customCf {
		kvStore.cf[name] = cfg
	}
	cf := cfSlice(kvStore.cf)

	if err := kvStore.OpenDbColumnFamilies(cf, opts.path); err != nil {
		return nil, err
	}

	addToPathDbMap(opts.path, kvStore)

	return kvStore, nil
}

func (opts RocksDBOpts) newRocksDBWithOptions(name string, dir string) (kv.RwDB, error) {
	dbPath := filepath.Join(dir, name+".db")
	db, err := grocksdb.OpenDb(opts, dbPath)
	if err != nil {
		return nil, err
	}
	ro := grocksdb.NewDefaultReadOptions()
	wo := grocksdb.NewDefaultWriteOptions()
	woSync := grocksdb.NewDefaultWriteOptions()
	woSync.SetSync(true)
	database := &RocksDB{
		db:     db,
		ro:     ro,
		wo:     wo,
		woSync: woSync,
	}
	return database, nil
}

func (opts RocksDBOpts) GetLabel() kv.Label { return opts.label }

func (opts RocksDBOpts) DBVerbosity(v kv.DBVerbosityLvl) RocksDBOpts {
	opts.verbosity = v
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

func (opts RocksDBOpts) TableCfgFunc(cf TableCfgFunc) RocksDBOpts {
	opts.cfConfig = cf
	return opts
}

func (opts RocksDBOpts) BlockSize(size int64) RocksDBOpts {
	opts.blockSize = size
	return opts
}

func (opts RocksDBOpts) BlockCache(size int64) RocksDBOpts {
	opts.blockCache = size
	return opts
}

func (opts RocksDBOpts) Statistics(enable bool) RocksDBOpts {
	opts.statistics = enable
	return opts
}

func (opts RocksDBOpts) MaxOpenFiles(max int64) RocksDBOpts {
	opts.maxOpenFiles = max
	return opts
}

func (opts RocksDBOpts) MmapRead(enable bool) RocksDBOpts {
	opts.mmapRead = enable
	return opts
}

func (opts RocksDBOpts) MmapWrite(enable bool) RocksDBOpts {
	opts.mmapWrite = enable
	return opts
}

func (opts RocksDBOpts) UnorderedWrite(enable bool) RocksDBOpts {
	opts.unorderedWrite = enable
	return opts
}

func (opts RocksDBOpts) PipelinedWrite(enable bool) RocksDBOpts {
	opts.pipelinedWrite = enable
	return opts
}
