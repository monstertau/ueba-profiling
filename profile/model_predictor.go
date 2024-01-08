package profile

import (
	"time"
	"ueba-profiling/model"
	"ueba-profiling/view"
)

type (
	FirstOccurrenceModel struct {
		attributes map[string]int
		startTime  time.Time
		endTime    time.Time
	}
	FirstOccurrencePredictor struct {
		id              string
		cfg             *view.ProfileConfig
		inChan          chan []byte
		bloom           *BloomFilterModel
		fullState       *CacheState[*FirstOccurrenceModel]
		communicationCh chan []*model.Frequency
	}
)

func NewFirstOccurrencePredictor(id string, cfg *view.ProfileConfig, inChan chan []byte, commCh chan []*model.Frequency) (*FirstOccurrencePredictor, error) {
	cacheState, err := NewCacheState[*FirstOccurrenceModel](1000, "./test_cache_db")
	if err != nil {
		return nil, err
	}
	return &FirstOccurrencePredictor{
		id:              id,
		cfg:             cfg,
		inChan:          inChan,
		fullState:       cacheState,
		communicationCh: commCh,
	}, nil
}

func (p *FirstOccurrencePredictor) Start() {
	go func() {
		for {
			select {
			case msg := <-p.inChan:
				p.predict(msg)
			case b := <-p.communicationCh:
				p.build(b)
			}
		}
	}()
}

func (p *FirstOccurrencePredictor) predict(msg []byte) {

}

func (p *FirstOccurrencePredictor) build(batches []*model.Frequency) {

}

func (p *FirstOccurrencePredictor) Stop() {

}
