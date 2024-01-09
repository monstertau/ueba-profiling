package profile

import (
	"github.com/pkg/errors"
	"strings"
	"time"
	"ueba-profiling/model"
	"ueba-profiling/util"
	"ueba-profiling/view"
)

type (
	FirstOccurrenceModel struct {
		occurrence int
		startTime  time.Time
		endTime    time.Time
	}
	FirstOccurrencePredictor struct {
		id                              string
		cfg                             *view.ProfileConfig
		entityGetters, attributeGetters []FieldGetter
		inChan                          chan []byte
		bloom                           *BloomFilterModel
		fullState                       *CacheState[*FirstOccurrenceModel]
		communicationCh                 chan []*model.Frequency
		outChan                         chan *model.Event
	}
)

func NewFirstOccurrencePredictor(id string, cfg *view.ProfileConfig, inChan chan []byte, outCh chan *model.Event, commCh chan []*model.Frequency) (*FirstOccurrencePredictor, error) {
	cacheState, err := NewCacheState[*FirstOccurrenceModel](1000, "./test_cache_db")
	if err != nil {
		return nil, err
	}
	var entityGetters, attributeGetters []FieldGetter
	for _, entity := range cfg.Entities {
		getter, err := fieldGetterFactory(entity)
		if err != nil {
			return nil, errors.Wrap(err, "in fieldGetterFactory")
		}
		entityGetters = append(entityGetters, getter)
	}
	for _, attribute := range cfg.Attributes {
		wrapper, err := fieldGetterFactory(attribute)
		if err != nil {
			return nil, errors.Wrap(err, "in fieldGetterFactory")
		}
		attributeGetters = append(attributeGetters, wrapper)
	}
	return &FirstOccurrencePredictor{
		id:               id,
		cfg:              cfg,
		entityGetters:    entityGetters,
		attributeGetters: attributeGetters,
		inChan:           inChan,
		outChan:          outCh,
		fullState:        cacheState,
		communicationCh:  commCh,
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
	normalizedEntity, ok := normalizeFieldValues(msg, p.entityGetters)
	if !ok {
		return
	}
	normalizedAttr, ok := normalizeFieldValues(msg, p.attributeGetters)
	if !ok {
		return
	}
	k := normalizedEntity + "_" + normalizedAttr
	if !p.bloom.Contains(k) {
		p.outChan <- &model.Event{
			RawEvent:   msg,
			Entities:   normalizedEntity,
			Attributes: normalizedAttr,
			Result:     1,
			Threshold:  p.cfg.Threshold,
		}
	}
	if _, ok := p.fullState.Get(k); !ok {
		p.outChan <- &model.Event{
			RawEvent:   msg,
			Entities:   normalizedEntity,
			Attributes: normalizedAttr,
			Result:     1,
			Threshold:  p.cfg.Threshold,
		}
	}
	p.bloom.Update(k)
}

func (p *FirstOccurrencePredictor) build(batches []*model.Frequency) {
	for _, b := range batches {
		if b.ID != p.id {
			continue
		}
		p.rebuildState(b)
	}
	keys := p.fullState.GetAllKey()
	p.bloom.Rebuild(keys)

}

func (p *FirstOccurrencePredictor) rebuildState(b *model.Frequency) {
	k := b.Entities + "_" + b.Attributes
	data, ok := p.fullState.Get(k)
	if !ok {
		return
	}
	data.occurrence += b.Count
	data.endTime = b.EndTime
	// TODO unmerge expired value
}

func (p *FirstOccurrencePredictor) Stop() {

}

func normalizeFieldValues(msg []byte, fieldGetters []FieldGetter) (string, bool) {
	fieldValues := make([]string, len(fieldGetters), len(fieldGetters))
	for i, fieldGetter := range fieldGetters {
		val, ok := fieldGetter.Get(msg)
		if !ok {
			return "", false
		}
		fieldValues[i], ok = util.ParseStringExtended(val)
		if !ok {
			return "", false
		}
	}
	normalizedTokens := strings.Join(fieldValues, "_")
	return normalizedTokens, true
}
