package store

import (
	"github.com/dgraph-io/badger/v4"
)

type StoreConfig struct {
	Path string
}

type StoreItem struct {
	Key   []byte
	Value []byte
}

type Store interface {
	Get([]byte) ([]byte, error)
	Set(StoreItem) error
	Del([]byte) error
	All([]byte, uint64) ([]StoreItem, error)

	Start() error
	Stop() error
}

func NewStore(config StoreConfig) Store {
	return &store{config: config}
}

type store struct {
	db     *badger.DB
	config StoreConfig
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

func (st *store) Set(item StoreItem) error {
	return st.db.Update(func(txn *badger.Txn) error {
		return txn.Set(item.Key, item.Value)
	})
}

func (st *store) Del(key []byte) error {
	return st.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func (st *store) All(prefix []byte, count uint64) ([]StoreItem, error) {
	result := []StoreItem{}
	return result, st.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		var err error
		for it.Seek(prefix); err == nil && it.ValidForPrefix(prefix) && count > 0; it.Next() {
			item := it.Item()

			var value []byte
			if value, err = item.ValueCopy(make([]byte, item.ValueSize())); err == nil {
				result = append(result, StoreItem{Key: item.Key(), Value: value})
			}
		}
		return err
	})
}

func (st *store) Start() error {
	if st.db == nil {
		db, err := badger.Open(badger.DefaultOptions(st.config.Path))
		if err == nil {
			st.db = db
		}
		return err
	}
	return nil
}

func (st *store) Stop() error {
	if st.db != nil {
		return st.db.Close()
	}
	return nil
}
