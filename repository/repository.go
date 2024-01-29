package repository

import (
	"time"
	"ueba-profiling/model"
)

type (
	IRepository interface {
		Persist(b []*model.Frequency) error
		GetFrequenciesFromRange(from time.Time, to time.Time) ([]*model.Frequency, error)
	}
	MongoRepository struct {
	}
)

func (m *MongoRepository) Persist(b []*model.Frequency) error {
	//TODO implement me
	return nil
}

func (m *MongoRepository) GetFrequenciesFromRange(from time.Time, to time.Time) ([]*model.Frequency, error) {
	//TODO implement me
	return nil, nil
}
