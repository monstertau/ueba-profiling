package profile

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"time"
	"ueba-profiling/model"
	"ueba-profiling/repository"
	"ueba-profiling/view"
)

type FirstOccurrenceBuilder struct {
	id              string
	cfg             *view.ProfileConfig
	inChan          chan []byte
	batch           []*model.Frequency
	saveDuration    time.Duration
	communicationCh chan []*model.Frequency
	repo            repository.IRepository
	logger          *logrus.Entry
}

func NewFirstOccurrenceBuilder(id string, inChan chan []byte, conf *view.ProfileConfig, repo repository.IRepository) (*FirstOccurrenceBuilder, error) {
	return &FirstOccurrenceBuilder{
		id:           id,
		cfg:          conf,
		inChan:       inChan,
		batch:        make([]*model.Frequency, 0),
		saveDuration: 5 * time.Second,
		repo:         repo,
		logger:       logrus.WithField("service", "builder"),
	}, nil
}

func (b *FirstOccurrenceBuilder) Start() {
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
				if len(b.batch) < 1000 {
					continue
				}
				b.dump()
			case <-ticker.C:
				if len(b.batch) == 0 {
					continue
				}
				b.dump()
			}
		}
	}()
}

func (b *FirstOccurrenceBuilder) dump() {
	err := b.repo.Persist(b.batch)
	if err != nil {
		b.logger.Errorf("error in persist frequency model: %v", err)
		return
	}
	b.communicationCh <- b.batch
	b.batch = make([]*model.Frequency, 0)
}

func (b *FirstOccurrenceBuilder) Stop() {

}

func (b *FirstOccurrenceBuilder) GetCommunicationChan() chan []*model.Frequency {
	return b.communicationCh
}
