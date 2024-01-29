package profile

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"os"
)

type (
	SerializerFunc[T any]           func(v T) ([]byte, error)
	DeserializerFunc[T any]         func(v []byte) (T, error)
	LevelDBCachingRepository[T any] struct {
		filePath     string
		db           *leveldb.DB
		serializer   SerializerFunc[T]
		deserializer DeserializerFunc[T]
		transaction  *leveldb.Transaction
	}
	Pair[T any] struct {
		key   string
		value T
	}
)

func NewLevelDBCachingRepository[T any](filePath string) (*LevelDBCachingRepository[T], error) {

	db, err := leveldb.OpenFile(filePath, &opt.Options{OpenFilesCacheCapacity: 100})
	if err != nil {
		return nil, err
	}
	serializerFunc := func(v T) ([]byte, error) {
		return json.Marshal(v)
	}
	deserializerFunc := func(v []byte) (T, error) {
		var res T
		err := json.Unmarshal(v, &res)
		return res, err
	}
	return &LevelDBCachingRepository[T]{
		filePath:     filePath,
		db:           db,
		serializer:   serializerFunc,
		deserializer: deserializerFunc,
	}, nil
}

func (r *LevelDBCachingRepository[T]) Get(key string) (T, bool) {
	var v T
	value, err := r.db.Get([]byte(key), &opt.ReadOptions{DontFillCache: true})
	if err != nil {
		return v, false
	}
	v, err = r.deserializer(value)
	if err != nil {
		return v, false
	}
	return v, true
}

func (r *LevelDBCachingRepository[T]) Set(key string, value T) error {
	v, err := r.serializer(value)
	if err != nil {
		return err
	}
	err = r.db.Put([]byte(key), v, nil)
	if err != nil {
		return err
	}
	return nil
}

func (r *LevelDBCachingRepository[T]) BulkUpdate(pairs []*Pair[T]) error {
	b := &leveldb.Batch{}
	for _, pair := range pairs {
		byteV, err := r.serializer(pair.value)
		if err != nil {
			continue
		}
		b.Put([]byte(pair.key), byteV)
	}
	return r.db.Write(b, nil)
}

func (r *LevelDBCachingRepository[T]) NewTransaction() error {
	tr, err := r.db.OpenTransaction()
	if err != nil {
		return err
	}
	r.transaction = tr
	return nil
}

func (r *LevelDBCachingRepository[T]) PrepareSet(key string, value T) error {
	v, err := r.serializer(value)
	if err != nil {
		return err
	}
	err = r.transaction.Put([]byte(key), v, nil)
	if err != nil {
		return err
	}
	return nil
}

func (r *LevelDBCachingRepository[T]) PrepareDelete(key string) error {
	err := r.transaction.Delete([]byte(key), nil)
	if err != nil {
		return err
	}
	return nil
}

func (r *LevelDBCachingRepository[T]) CommitTransaction() error {
	return r.transaction.Commit()
}

func (r *LevelDBCachingRepository[T]) TotalKey() int {
	count := 0
	iter := r.db.NewIterator(nil, &opt.ReadOptions{DontFillCache: true})
	for iter.Next() {
		count += 1
	}
	iter.Release()
	_ = iter.Error()
	return count
}

func (r *LevelDBCachingRepository[T]) Iterate() chan *Pair[T] {
	pChan := make(chan *Pair[T])
	go func() {
		iter := r.db.NewIterator(nil, &opt.ReadOptions{DontFillCache: true})
		for iter.Next() {
			key := iter.Key()
			value := iter.Value()
			valueT, err := r.deserializer(value)
			if err != nil {
				continue
			}
			pChan <- &Pair[T]{key: string(key), value: valueT}
		}
		iter.Release()
		_ = iter.Error()
		close(pChan)
	}()
	return pChan
}

func (r *LevelDBCachingRepository[T]) CleanUp() error {
	err := r.db.Close()
	if err != nil {
		return err
	}
	return os.RemoveAll(r.filePath)
}

func (r *LevelDBCachingRepository[T]) CopyFrom(another *LevelDBCachingRepository[T]) error {
	iter := another.db.NewIterator(nil, &opt.ReadOptions{DontFillCache: true})
	defer iter.Release()
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		err := r.db.Put(key, value, nil)
		if err != nil {
			return errors.Wrap(err, "cannot put key-value to level db")
		}
	}
	return iter.Error()
}
