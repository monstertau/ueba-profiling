package profile

type (
	FirstOccurrencePredictor struct {
		id        string
		threshold float64
		inChan    chan []byte
		bloom     *BloomFilterModel
		fullState *CacheState[map[string]int]
	}
)

func NewFirstOccurrencePredictor(id string, threshold float64, inChan chan []byte) (*FirstOccurrencePredictor, error) {
	cacheState, err := NewCacheState[map[string]int](1000, "./test_cache_db")
	if err != nil {
		return nil, err
	}
	return &FirstOccurrencePredictor{
		id:        id,
		threshold: threshold,
		inChan:    inChan,
		fullState: cacheState,
	}, nil
}

func (*FirstOccurrencePredictor) Start() {

}

func (*FirstOccurrencePredictor) Stop() {

}
