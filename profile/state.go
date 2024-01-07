package profile

import (
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/syndtr/goleveldb/leveldb"
)

type CacheState[T any] struct {
	cache   *lru.Cache[string, T]
	cacheDB *leveldb.DB
}

func NewCacheState[T any](maxSize int, cachingPath string) (*CacheState[T], error) {
	l, err := lru.New[string, T](maxSize)
	if err != nil {
		return nil, err
	}
	db, err := leveldb.OpenFile(cachingPath, nil)
	if err != nil {
		return nil, err
	}
	return &CacheState[T]{
		cache:   l,
		cacheDB: db,
	}, nil

}
