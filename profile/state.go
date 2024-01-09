package profile

import (
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type (
	CacheState[T any] struct {
		sync.RWMutex
		PID             string
		diskCachingRepo *LevelDBCachingRepository[T]
		MaxSize         int
		LRUState        *lru.Cache[string, T]
		logger          *logrus.Entry
		syncDuration    time.Duration
		managerChan     chan interface{}
	}
	FirstOccurrenceState struct {
		CacheState[map[string]int]
	}
)

func NewCacheState[T any](maxSize int, cachingPath string) (*CacheState[T], error) {
	return nil, nil
}

func (s *CacheState[T]) Get(key string) (T, bool) {
	var t T
	return t, false
}

func (s *CacheState[T]) Set(key string, value T) bool {
	return false
}

func (s *CacheState[T]) GetAllKey() []string {
	return nil
}
