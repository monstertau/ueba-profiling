package profile

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sort"
	"strings"
	"time"
	"ueba-profiling/model"
	"ueba-profiling/repository"
	"ueba-profiling/util"
	"ueba-profiling/view"
)

type (
	IPredictor interface {
		Start()
		Predict(msg []byte)
		Stop()
	}
	FirstOccurrenceModel struct {
		Occurrence int
		StartTime  time.Time
		EndTime    time.Time
	}
	FirstOccurrencePredictor struct {
		id                              string
		cfg                             *view.ProfileConfig
		profileDuration                 time.Duration
		entityGetters, attributeGetters []FieldGetter
		inChan                          chan []byte
		bloom                           *BloomFilterModel
		fullState                       *CacheState[*FirstOccurrenceModel]
		communicationCh                 chan []*model.Frequency
		outChan                         chan *model.Event
		repo                            repository.IRepository
		logger                          *logrus.Entry
	}
)

func FactoryPredictor(jobConfig *view.JobConfig, repo repository.IRepository,
	inChan chan []byte, outCh chan *model.Event, commCh chan []*model.Frequency) (IPredictor, error) {
	switch jobConfig.ProfileConfig.ProfileType {
	case profileTypeFirstOccurrence:
		return NewFirstOccurrencePredictor(jobConfig.ID, jobConfig.ProfileConfig, repo, inChan, outCh, commCh)
	case profileTypeRarity:
		return NewRarityPredictor(jobConfig.ID, jobConfig.ProfileConfig, repo, inChan, outCh, commCh)
	default:
		return nil, fmt.Errorf("cant find profile with type %v", jobConfig.ProfileConfig.ProfileType)
	}
}

func NewFirstOccurrencePredictor(
	id string,
	cfg *view.ProfileConfig,
	repo repository.IRepository,
	inChan chan []byte,
	outCh chan *model.Event,
	commCh chan []*model.Frequency,
) (*FirstOccurrencePredictor, error) {

	cacheState, err := NewCacheState[*FirstOccurrenceModel](id, 1000, "./test_cache_db")
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
	profileDuration, err := util.ParseDurationExtended(cfg.ProfileTime)
	if err != nil {
		return nil, errors.Wrap(err, "parse profile_time param")
	}
	return &FirstOccurrencePredictor{
		id:               id,
		cfg:              cfg,
		profileDuration:  profileDuration,
		entityGetters:    entityGetters,
		attributeGetters: attributeGetters,
		inChan:           inChan,
		outChan:          outCh,
		fullState:        cacheState,
		communicationCh:  commCh,
		repo:             repo,
		logger:           logrus.WithField("first_occurrence_predictor", id),
	}, nil
}

func (p *FirstOccurrencePredictor) Start() {
	go func() {
		for {
			select {
			case msg := <-p.inChan:
				p.Predict(msg)
			case b := <-p.communicationCh:
				p.build(b)
			}
		}
	}()
}

func (p *FirstOccurrencePredictor) Predict(msg []byte) {
	normalizedEntity, ok := normalizeFieldValues(msg, p.entityGetters)
	if !ok {
		return
	}
	normalizedAttr, ok := normalizeFieldValues(msg, p.attributeGetters)
	if !ok {
		return
	}
	k := normalizedEntity + "_" + normalizedAttr
	//if !p.bloom.Contains(k) {
	// p.outChan <- &model.Event{
	// RawEvent: msg,
	// Entities: normalizedEntity,
	// Attributes: normalizedAttr,
	// Result: 1,
	// Threshold: p.cfg.Threshold,
	// }
	//}
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
		//if b.ID != p.id {
		// continue
		//}
		p.rebuildState(b)
	}
	keys := p.fullState.GetAllKey()
	p.bloom.Rebuild(keys)
}

func (p *FirstOccurrencePredictor) rebuildState(b *model.Frequency) {
	k := b.Entities + "_" + b.Attributes
	data, ok := p.fullState.Get(k)
	if !ok {
		data = &FirstOccurrenceModel{
			Occurrence: 0,
			StartTime:  b.EndTime.Time.Add(-p.profileDuration),
			EndTime:    b.EndTime.Time,
		}
	}
	defer p.fullState.Set(k, data)

	data.Occurrence += b.Count
	data.EndTime = b.EndTime.Time
	// TODO unmerge expired Value
	alignedStartTime := data.EndTime.Add(-p.profileDuration)
	freqs, err := p.repo.GetFrequenciesFromRange(data.StartTime, alignedStartTime)
	if err != nil {
		p.logger.Errorf("error in GetFrequenciesFromRange:%v", err)
		return
	}
	for _, f := range freqs {
		data.Occurrence -= f.Count
	}
	data.StartTime = alignedStartTime
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

type (
	Statistic struct {
		StartTime    time.Time
		EndTime      time.Time
		PredictedVal float64
		Count        int64
	}
	RarityModel struct {
		Htg map[string]*Statistic
		Sum int64
	}
	RarityPredictor struct {
		id                              string
		cfg                             *view.ProfileConfig
		profileDuration                 time.Duration
		entityGetters, attributeGetters []FieldGetter
		inChan                          chan []byte
		fullState                       *CacheState[*RarityModel]
		communicationCh                 chan []*model.Frequency
		outChan                         chan *model.Event
		repo                            repository.IRepository
		logger                          *logrus.Entry
	}
)

func NewRarityPredictor(
	id string,
	cfg *view.ProfileConfig,
	repo repository.IRepository,
	inChan chan []byte,
	outCh chan *model.Event,
	commCh chan []*model.Frequency,
) (*RarityPredictor, error) {
	cacheState, err := NewCacheState[*RarityModel](id, 1000, "./test_cache_db")
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
	profileDuration, err := util.ParseDurationExtended(cfg.ProfileTime)
	if err != nil {
		return nil, errors.Wrap(err, "parse profile_time param")
	}
	return &RarityPredictor{
		id:               id,
		cfg:              cfg,
		profileDuration:  profileDuration,
		entityGetters:    entityGetters,
		attributeGetters: attributeGetters,
		inChan:           inChan,
		outChan:          outCh,
		fullState:        cacheState,
		communicationCh:  commCh,
		repo:             repo,
		logger:           logrus.WithField("rarity_predictor", id),
	}, nil
}

func (p *RarityPredictor) Start() {
	go func() {
		for {
			select {
			case msg := <-p.inChan:
				p.Predict(msg)
			case b := <-p.communicationCh:
				p.build(b)
			}
		}
	}()
}

func (p *RarityPredictor) Predict(msg []byte) {
	normalizedEntity, ok := normalizeFieldValues(msg, p.entityGetters)
	if !ok {
		return
	}
	normalizedAttr, ok := normalizeFieldValues(msg, p.attributeGetters)
	if !ok {
		return
	}
	m, ok := p.fullState.Get(normalizedEntity)
	if !ok { // entity not found, need further profile building
		return
	}
	v, ok := m.Htg[normalizedAttr]
	if !ok { // attribute not found, need further profile building
		return
	}
	if v.PredictedVal >= p.cfg.Threshold {
		p.outChan <- &model.Event{
			RawEvent:   msg,
			Entities:   normalizedEntity,
			Attributes: normalizedAttr,
			Result:     v.PredictedVal,
			Threshold:  p.cfg.Threshold,
		}
	}

}

func (p *RarityPredictor) build(batches []*model.Frequency) {
	for _, b := range batches {
		//if b.ID != p.id {
		// continue
		//}
		p.rebuildState(b)
	}
	// pre-Predict
	for _, kv := range p.fullState.GetAll() {
		m := kv.Value
		m.PrePredict()
		p.fullState.Set(kv.Key, m)
	}
}

func (p *RarityPredictor) rebuildState(b *model.Frequency) {
	data, ok := p.fullState.Get(b.Entities)
	if !ok {
		data = &RarityModel{
			Htg: make(map[string]*Statistic),
		}

	}
	defer p.fullState.Set(b.Entities, data)
	stats, ok := data.Htg[b.Attributes]
	if !ok {
		stats = &Statistic{
			StartTime: b.EndTime.Time.Add(-p.profileDuration),
			EndTime:   b.EndTime.Time,
		}
		data.Htg[b.Attributes] = stats
	}
	stats.Count += 1
	stats.EndTime = b.EndTime.Time
	data.Sum += 1

	// TODO unmerge expired Value
	alignedStartTime := stats.EndTime.Add(-p.profileDuration)
	freqs, err := p.repo.GetFrequenciesFromRange(stats.StartTime, alignedStartTime)
	if err != nil {
		p.logger.Errorf("error in GetFrequenciesFromRange:%v", err)
		return
	}
	for range freqs {
		stats.Count -= 1
		data.Sum -= 1
	}
	stats.StartTime = alignedStartTime
}

func (p *RarityPredictor) Stop() {

}

func (m *RarityModel) PrePredict() {
	nVal := len(m.Htg)
	if nVal == 0 {
		return
	}
	listVal := make([]int64, 0, nVal)
	for _, val := range m.Htg {
		listVal = append(listVal, val.Count)
	}
	sort.Slice(listVal, func(i, j int) bool {
		return listVal[i] < listVal[j]
	})
	accumulateVal := make([]int64, nVal, nVal)
	accumulateVal[0] = listVal[0]
	for i := 1; i < nVal; i++ {
		accumulateVal[i] = accumulateVal[i-1] + listVal[i]
	}
	for fea := range m.Htg {
		m.Htg[fea].PredictedVal = m.RawPredict(fea, listVal, accumulateVal)
	}
}

func (m *RarityModel) RawPredict(feature string, listVal, accumulateVal []int64) float64 {
	count := m.Htg[feature].Count
	threshold := 2 * count
	if id := getLowerBoundThreshold(threshold, listVal); id != -1 {
		count += accumulateVal[id]
		if m.Htg[feature].Count <= threshold {
			count -= m.Htg[feature].Count
		}
	}

	pro := float64(1 - count/m.Sum)
	if pro < 0 {
		pro = 0
	}
	return pro
}

func getLowerBoundThreshold(threshold int64, listVal []int64) int {
	var (
		first = 0
		last  = len(listVal) - 1
		mid   = 0
		id    = -1
	)

	for first <= last {
		mid = (first + last) / 2
		if listVal[mid] <= threshold {
			id = mid
			first = mid + 1
		} else {
			last = mid - 1
		}
	}
	return id
}
