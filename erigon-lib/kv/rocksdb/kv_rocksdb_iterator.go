package rocksdb

import (
	"bytes"

	"github.com/linxGnu/grocksdb"
)

type RocksCursor struct {
	tx *RocksTx

	itr *grocksdb.Iterator
}

func (r RocksCursor) First() ([]byte, []byte, error) {
	//TODO implement me
	panic("implement me - First")
}

func (r RocksCursor) Seek(seek []byte) ([]byte, []byte, error) {
	//TODO implement me
	panic("implement me - Seek")
}

func (r RocksCursor) SeekExact(key []byte) ([]byte, []byte, error) {
	//TODO implement me
	panic("implement me - SeekExact")
}

func (r RocksCursor) Next() ([]byte, []byte, error) {
	//TODO implement me
	panic("implement me - Next")
}

func (r RocksCursor) Prev() ([]byte, []byte, error) {
	//TODO implement me
	panic("implement me - Prev")
}

func (r RocksCursor) Last() ([]byte, []byte, error) {
	r.itr.SeekToLast()
	if r.itr.Valid() {
		return bytes.Clone(r.itr.Key().Data()), bytes.Clone(r.itr.Value().Data()), nil
	}
	return nil, nil, nil
}

func (r RocksCursor) Current() ([]byte, []byte, error) {
	//TODO implement me
	panic("implement me - Current")
}

func (r RocksCursor) Count() (uint64, error) {
	//TODO implement me
	panic("implement me - Count")
}

func (r RocksCursor) Close() {
	if r.itr != nil {
		r.itr.Close()
	}
}

func (r RocksCursor) Put(k, v []byte) error {
	//TODO implement me
	panic("implement me - RocksDB iterator doesnt support Put!")
}

func (r RocksCursor) Append(k []byte, v []byte) error {
	//TODO implement me
	panic("implement me - Append")
}

func (r RocksCursor) Delete(k []byte) error {
	//TODO implement me
	panic("implement me - Delete")
}

func (r RocksCursor) DeleteCurrent() error {
	//TODO implement me
	panic("implement me - DeleteCurrent")
}
