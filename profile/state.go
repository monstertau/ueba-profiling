package profile

import (
	"fmt"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"math"
	"os"
	"sync"
)

type (
	CacheState[T any] struct {
		sync.RWMutex
		PID             string
		diskCachingRepo *LevelDBCachingRepository[T]
		MaxSize         int
		LRUState        *lru.Cache[string, T]
		logger          *logrus.Entry
		//syncDuration    time.Duration
		//managerChan     chan interface{}
	}
)

func NewCacheState[T any](id string, maxSize int, cachingPath string) (*CacheState[T], error) {
	// init disk cache
	_ = os.MkdirAll(cachingPath, os.ModePerm)
	filePath, err := os.MkdirTemp(cachingPath, fmt.Sprintf("first-occurrence-%v-*.db", id))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create temp directory")
	}
	cachingRepo, err := NewLevelDBCachingRepository[T](filePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create db caching repo")
	}
	// init memory cache
	lruState, err := lru.New[string, T](math.MaxInt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create lru state")
	}
	return &CacheState[T]{
		PID:             id,
		diskCachingRepo: cachingRepo,
		MaxSize:         maxSize,
		LRUState:        lruState,
		logger:          logrus.WithField("layer", "state"),
	}, nil
}

func (s *CacheState[T]) Get(key string) (T, bool) {
	v, ok := s.LRUState.Get(key)
	if !ok {
		return s.GetAndUpdateFromDisk(key)
	}
	return v, true
}

func (s *CacheState[T]) GetAndUpdateFromDisk(key string) (T, bool) {
	s.Lock()
	defer s.Unlock()
	v, ok := s.diskCachingRepo.Get(key)
	if !ok {
		var t T
		return t, false
	}
	s.updateInFlightCache(key, v)
	return v, ok
}

func (s *CacheState[T]) updateInFlightCache(key string, value T) {
	if s.LRUState.Len() >= s.MaxSize {
		oldestEnt, oldestAttMap, _ := s.LRUState.RemoveOldest()
		err := s.diskCachingRepo.Set(oldestEnt, oldestAttMap)
		if err != nil {
			s.logger.Warnf("error in append old data to disk cache:%v", err)
		}
	}
	s.LRUState.Add(key, value)
}

func (s *CacheState[T]) Set(key string, value T) bool {
	s.LRUState.Add(key, value)
	return true
}

func (s *CacheState[T]) GetAllKey() []string {
	return nil
}
