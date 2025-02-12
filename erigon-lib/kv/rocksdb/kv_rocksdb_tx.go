package rocksdb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/iter"
	"github.com/ledgerwatch/erigon-lib/kv/order"
	"github.com/linxGnu/grocksdb"
	"unsafe"
)

/*Not actually using RocksDB Tx just implementing interface*/

type RocksTx struct {
	db *RocksKV

	batch *grocksdb.WriteBatch
}

func NewRocksDBBatch(db *RocksKV) *RocksTx {
	return &RocksTx{
		db:    db,
		batch: grocksdb.NewWriteBatch(),
	}
}

func (b *RocksTx) assertOpen() {
	if b.batch == nil {
		panic("batch has been written or closed")
	}
}

func (rtx *RocksTx) Has(table string, key []byte) (bool, error) {
	iterator := rtx.batch.NewIterator()
	if err := iterator.Record(); err != nil {
		return false, fmt.Errorf("iterate batch: %w", err)
	}
	for iterator.Next() {
		if bytes.Compare(iterator.Record().Key, key) == 0 {
			return true, nil
		}
	}

	return true, nil

}

func (rtx *RocksTx) GetOne(table string, key []byte) (val []byte, err error) {
	iterator := rtx.batch.NewIterator()
	if err := iterator.Record(); err != nil {
		return nil, fmt.Errorf("iterate batch: %w", err)
	}
	for iterator.Next() {
		if bytes.Compare(iterator.Record().Key, key) == 0 {
			return iterator.Record().Value, nil
		}
	}

	return nil, nil
}

func (rtx *RocksTx) ForEach(table string, fromPrefix []byte, walker func(k []byte, v []byte) error) error {
	panic("implement me - ForEach")
	return nil
}

func (rtx *RocksTx) ForPrefix(table string, prefix []byte, walker func(k []byte, v []byte) error) error {
	//TODO implement me
	panic("implement me - ForPrefix")
}

func (rtx *RocksTx) ForAmount(table string, prefix []byte, amount uint32, walker func(k []byte, v []byte) error) error {
	//TODO implement me
	panic("implement me - ForAmount")
}

func (rtx *RocksTx) Commit() error {
	rtx.assertOpen()
	err := rtx.db.db.Write(rtx.db.wo, rtx.batch)
	if err != nil {
		return err
	}
	rtx.Close()
	return nil
}

// Close implements Batch.
func (rtx *RocksTx) Close() {
	if rtx.batch != nil {
		rtx.batch.Destroy()
		rtx.batch = nil
	}
}

func (rtx *RocksTx) Rollback() {
	if rtx.batch != nil {
		rtx.batch.Clear()
	}
	return
}

func (rtx *RocksTx) ListBuckets() ([]string, error) {
	//TODO implement me
	panic("implement me- ListBuckets")
}

func (rtx *RocksTx) ViewID() uint64 {
	//TODO implement me
	panic("implement me - ViewID")
}

func (rtx *RocksTx) Cursor(table string) (kv.Cursor, error) {
	return rtx.RwCursor(table)
}

func (rtx *RocksTx) CursorDupSort(table string) (kv.CursorDupSort, error) {
	//TODO implement me
	panic("implement me - CursorDupSort")
}

func (rtx *RocksTx) DBSize() (uint64, error) {
	//TODO implement me
	panic("implement me - DBSize")
}

func (rtx *RocksTx) Range(table string, fromPrefix, toPrefix []byte) (iter.KV, error) {
	//TODO implement me
	panic("implement me - Range")
}

func (rtx *RocksTx) RangeAscend(table string, fromPrefix, toPrefix []byte, limit int) (iter.KV, error) {
	//TODO implement me
	panic("implement me - RangeAscend")
}

func (rtx *RocksTx) RangeDescend(table string, fromPrefix, toPrefix []byte, limit int) (iter.KV, error) {
	//TODO implement me
	panic("implement me - RangeDescend")
}

func (rtx *RocksTx) Prefix(table string, prefix []byte) (iter.KV, error) {
	//TODO implement me
	panic("implement me - Prefix")
}

func (rtx *RocksTx) RangeDupSort(table string, key []byte, fromPrefix, toPrefix []byte, asc order.By, limit int) (iter.KV, error) {
	//TODO implement me
	panic("implement me - RangeDupSort")
}

func (rtx *RocksTx) CHandle() unsafe.Pointer {
	//TODO implement me
	panic("implement me - CHandle")
}

func (rtx *RocksTx) BucketSize(table string) (uint64, error) {
	//TODO implement me
	panic("implement me - BucketSize")
}

func (rtx *RocksTx) Put(table string, k, v []byte) error {
	rtx.assertOpen()
	rtx.batch.Put(k, v)
	return nil
}

func (rtx *RocksTx) Delete(table string, k []byte) error {
	rtx.assertOpen()
	rtx.batch.Delete(k)
	return nil
}

func (rtx *RocksTx) ReadSequence(table string) (uint64, error) {
	val, err := rtx.GetOne(kv.Sequence, []byte(table))
	if err != nil {
		return 0, err
	}

	var currentV uint64
	if len(val) > 0 {
		currentV = binary.BigEndian.Uint64(val)
	}
	return currentV, nil
}

func (rtx *RocksTx) IncrementSequence(table string, amount uint64) (uint64, error) {
	val, err := rtx.GetOne(kv.Sequence, []byte(table))
	if err != nil {
		return 0, err
	}

	var currentV uint64 = 0
	if len(val) > 0 {
		currentV = binary.BigEndian.Uint64(val)
	}
	newVBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(newVBytes, currentV+amount)

	err = rtx.Put(kv.Sequence, []byte(table), newVBytes)
	if err != nil {
		return 0, err
	}
	return currentV, nil
}

func (rtx *RocksTx) Append(table string, k, v []byte) error {
	//TODO implement me
	panic("implement me - Append")
}

func (rtx *RocksTx) AppendDup(table string, k, v []byte) error {
	//TODO implement me
	panic("implement me - AppendDup")
}

func (rtx *RocksTx) DropBucket(s string) error {
	panic("implement me - DropBucket")
	return nil
}

func (rtx *RocksTx) CreateBucket(name string) error {
	panic("implement me - CreateBucket")
	return nil
}

func (rtx *RocksTx) ExistsBucket(s string) (bool, error) {
	//TODO implement me
	panic("implement me - ExistsBucket")
}

func (rtx *RocksTx) ClearBucket(s string) error {
	//TODO fix to clear instead of drop
	return rtx.DropBucket(s)

}

func (rtx *RocksTx) RwCursor(table string) (kv.RwCursor, error) {
	return rtx.stdCursor(table)
}

func (rtx *RocksTx) stdCursor(table string) (kv.RwCursor, error) {
	itr := rtx.db.db.NewIterator(rtx.db.ro)
	c := &RocksCursor{
		tx:  rtx,
		itr: itr,
	}
	return c, nil
}

func (rtx *RocksTx) RwCursorDupSort(table string) (kv.RwCursorDupSort, error) {
	//TODO implement me
	panic("implement me - RwCursorDupSort")
}

func (rtx *RocksTx) CollectMetrics() {
	//TODO implement me
	panic("implement me - CollectMetrics")
}
