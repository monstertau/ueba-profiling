package repository

import "ueba-profiling/model"

type (
	IRepository interface {
		Persist(b []*model.Frequency) error
	}
	MongoRepository struct {
	}
)

func (m *MongoRepository) Persist(b []*model.Frequency) error {
	//TODO implement me
	panic("implement me")
}
