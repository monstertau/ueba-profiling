package profile

import (
	"fmt"
	"time"
	"ueba-profiling/config"
	"ueba-profiling/external"
	"ueba-profiling/repository"
	"ueba-profiling/sink"
	"ueba-profiling/view"
)

const (
	profileTypeFirstOccurrence = "first_occurrence"
)

type (
	IWorker interface {
		Start()
		Stop()
	}
	FirstOccurrenceWorker struct {
		id                                 string
		builder                            *FirstOccurrenceBuilder
		builderConsumer, predictorConsumer external.KafkaConsumer
		predictor                          *FirstOccurrencePredictor
		sinkWorker                         *sink.SinkWorker
	}
)

func NewProfilingWorker(jobConfig *view.JobConfig) (IWorker, error) {
	switch jobConfig.ProfileConfig.ProfileType {
	case profileTypeFirstOccurrence:
		return NewFirstOccurrenceWorker(jobConfig, &repository.MongoRepository{})
	default:
		return nil, fmt.Errorf("cant find profile with type %v", jobConfig.ProfileConfig.ProfileType)
	}
}

func NewFirstOccurrenceWorker(jobConfig *view.JobConfig, repo repository.IRepository) (*FirstOccurrenceWorker, error) {
	builderCfg := map[string]interface{}{
		"bootstrap.servers": jobConfig.LogSourceBuilderConfig.BootstrapServer,
		"topics":            jobConfig.LogSourceBuilderConfig.Topic,
		"group.id":          fmt.Sprintf("%s_%v", config.AppConfig.KafkaLogSource.GroupID, jobConfig.ID),
	}
	builderConsumer, err := external.FactoryConsumer(builderCfg)
	if err != nil {
		return nil, err
	}
	builder, err := NewFirstOccurrenceBuilder(jobConfig.ID, builderConsumer.RcvChannel(), jobConfig.ProfileConfig, repo)
	if err != nil {
		return nil, err
	}

	predictorCfg := map[string]interface{}{
		"bootstrap.servers": jobConfig.LogSourcePredictorConfig.BootstrapServer,
		"topics":            jobConfig.LogSourcePredictorConfig.Topic,
		"group.id":          fmt.Sprintf("%s_%v", config.AppConfig.KafkaLogSource.GroupID, jobConfig.ID),
	}
	predictorConsumer, err := external.FactoryConsumer(predictorCfg)
	if err != nil {
		return nil, err
	}
	sinkWorker, err := sink.NewSinkWorker(jobConfig.OutputConfig)
	if err != nil {
		return nil, err
	}
	predictor, err := NewFirstOccurrencePredictor(
		jobConfig.ID,
		jobConfig.ProfileConfig,
		repo,
		predictorConsumer.RcvChannel(),
		sinkWorker.GetInChan(),
		builder.GetCommunicationChan())
	if err != nil {
		return nil, err
	}

	return &FirstOccurrenceWorker{
		id:                jobConfig.ID,
		builder:           builder,
		builderConsumer:   builderConsumer,
		predictor:         predictor,
		predictorConsumer: predictorConsumer,
	}, nil
}

func (w *FirstOccurrenceWorker) Start() {
	w.builder.Start()
	w.builderConsumer.Start()
	time.Sleep(10 * time.Second)
	w.predictorConsumer.Start()
	w.predictor.Start()
}

func (w *FirstOccurrenceWorker) Stop() {
	w.builderConsumer.Stop()
	w.predictorConsumer.Stop()
	w.builder.Stop()
	w.predictor.Stop()
}
