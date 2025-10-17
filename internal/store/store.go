package store

import (
	"io"
	netfs "netfs/internal"

	"github.com/dgraph-io/badger/v4"
)

type Store interface {
	Get([]byte) ([]byte, error)
	Set([]byte, []byte) error
	Del([]byte) error
	All([]byte, uint64) ([][]byte, error)

	io.Closer
}

func NewStore(config *netfs.DatabaseConfig) (Store, error) {
	db, err := badger.Open(badger.DefaultOptions(config.Path))
	if err == nil {
		return &store{db: db}, nil
	}
	return nil, err
}

type store struct {
	db *badger.DB
}

func (st *store) Get(key []byte) ([]byte, error) {
	var data []byte
	return data, st.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err == nil {
			err = item.Value(func(val []byte) error {
				data = val
				return nil
			})
		}
		return err
	})
}

func (st *store) Set(key []byte, value []byte) error {
	return st.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

func (st *store) Del(key []byte) error {
	return st.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func (st *store) All(prefix []byte, count uint64) ([][]byte, error) {
	result := [][]byte{}
	return result, st.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		var err error
		for it.Seek(prefix); err == nil && it.ValidForPrefix(prefix) && count > 0; it.Next() {
			item := it.Item()

			var data []byte
			if data, err = item.ValueCopy(make([]byte, item.ValueSize())); err == nil {
				result = append(result, data)
			}
		}
		return err
	})
}

func (st *store) Close() error {
	return st.db.Close()
}
