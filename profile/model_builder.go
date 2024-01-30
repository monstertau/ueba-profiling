package profile

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"time"
	"ueba-profiling/model"
	"ueba-profiling/repository"
	"ueba-profiling/view"
)

type (
	ModelBuilder struct {
		id              string
		cfg             *view.ProfileConfig
		inChan          chan []byte
		batch           []*model.Frequency
		saveDuration    time.Duration
		communicationCh chan []*model.Frequency
		repo            repository.IRepository
		logger          *logrus.Entry
	}
)

func NewModelBuilder(id string, inChan chan []byte, conf *view.ProfileConfig, repo repository.IRepository) (*ModelBuilder, error) {
	return &ModelBuilder{
		id:              id,
		cfg:             conf,
		inChan:          inChan,
		batch:           make([]*model.Frequency, 0),
		saveDuration:    5 * time.Second,
		repo:            repo,
		communicationCh: make(chan []*model.Frequency, 1000),
		logger:          logrus.WithField("service", "builder"),
	}, nil
}

func (b *ModelBuilder) Start() {
	go func() {
		ticker := time.NewTicker(b.saveDuration)
		for {
			select {
			case msg := <-b.inChan:
				var m *model.Frequency
				err := json.Unmarshal(msg, &m)
				if err != nil {
					b.logger.Errorf("error in unmarshal frequency model: %v", err)
					continue
				}
				b.batch = append(b.batch, m)
				if len(b.batch) > 1000 {
					b.dump()
				}
			case <-ticker.C:
				if len(b.batch) == 0 {
					continue
				}
				b.dump()
			}
		}
	}()
}

func (b *ModelBuilder) dump() {
	err := b.repo.Persist(b.batch)
	if err != nil {
		b.logger.Errorf("error in persist frequency model: %v", err)
		return
	}
	b.communicationCh <- b.batch
	b.batch = make([]*model.Frequency, 0)
}

func (b *ModelBuilder) Stop() {

}

func (b *ModelBuilder) GetCommunicationChan() chan []*model.Frequency {
	return b.communicationCh
}
