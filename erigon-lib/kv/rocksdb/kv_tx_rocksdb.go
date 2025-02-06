package rocksdb

import (
	"context"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/iter"
	"github.com/ledgerwatch/erigon-lib/kv/order"
	"github.com/linxGnu/grocksdb"
	"unsafe"
)

type RocksTx struct {
	id  uint64
	kv  *RocksKV
	ctx context.Context
	wo  *grocksdb.WriteOptions
	ro  *grocksdb.ReadOptions
}

func (r RocksTx) Has(table string, key []byte) (bool, error) {
	if cfHandle, exists := r.kv.cfHandles[table]; exists {
		psh, err := r.kv.db.GetPinnedCF(r.ro, cfHandle, key)
		if err != nil {
			return false, err
		}
		return psh.Exists(), nil
	} else {
		return false, nil
	}
}

func (r RocksTx) GetOne(table string, key []byte) (val []byte, err error) {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) ForEach(table string, fromPrefix []byte, walker func(k []byte, v []byte) error) error {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) ForPrefix(table string, prefix []byte, walker func(k []byte, v []byte) error) error {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) ForAmount(table string, prefix []byte, amount uint32, walker func(k []byte, v []byte) error) error {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) Commit() error {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) Rollback() {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) ReadSequence(table string) (uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) ListBuckets() ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) ViewID() uint64 {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) Cursor(table string) (kv.Cursor, error) {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) CursorDupSort(table string) (kv.CursorDupSort, error) {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) DBSize() (uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) Range(table string, fromPrefix, toPrefix []byte) (iter.KV, error) {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) RangeAscend(table string, fromPrefix, toPrefix []byte, limit int) (iter.KV, error) {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) RangeDescend(table string, fromPrefix, toPrefix []byte, limit int) (iter.KV, error) {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) Prefix(table string, prefix []byte) (iter.KV, error) {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) RangeDupSort(table string, key []byte, fromPrefix, toPrefix []byte, asc order.By, limit int) (iter.KV, error) {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) CHandle() unsafe.Pointer {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) BucketSize(table string) (uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) Put(table string, k, v []byte) error {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) Delete(table string, k []byte) error {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) IncrementSequence(table string, amount uint64) (uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) Append(table string, k, v []byte) error {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) AppendDup(table string, k, v []byte) error {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) DropBucket(s string) error {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) CreateBucket(s string) error {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) ExistsBucket(s string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) ClearBucket(s string) error {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) RwCursor(table string) (kv.RwCursor, error) {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) RwCursorDupSort(table string) (kv.RwCursorDupSort, error) {
	//TODO implement me
	panic("implement me")
}

func (r RocksTx) CollectMetrics() {
	//TODO implement me
	panic("implement me")
}
