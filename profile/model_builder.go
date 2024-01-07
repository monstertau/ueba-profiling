package profile

import (
	"time"
	"ueba-profiling/model"
	"ueba-profiling/repository"
	"ueba-profiling/view"
)

type FirstOccurrenceBuilder struct {
	id         string
	inChan     chan []byte
	batch      []*model.Frequency
	saveTicker time.Duration
	repo       repository.IRepository
}

func NewFirstOccurrenceBuilder(id string, inChan chan []byte, conf *view.ProfileConfig, repo repository.IRepository) (*FirstOccurrenceBuilder, error) {
	return &FirstOccurrenceBuilder{
		inChan:     inChan,
		batch:      make([]*model.Frequency, 100),
		saveTicker: 5 * time.Second,
		repo:       repo,
	}, nil
}

func (*FirstOccurrenceBuilder) Start() {

}

func (*FirstOccurrenceBuilder) Stop() {

}
