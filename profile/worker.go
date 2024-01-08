package profile

import (
	"fmt"
	"ueba-profiling/config"
	"ueba-profiling/external"
	"ueba-profiling/repository"
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
		id        string
		builder   *FirstOccurrenceBuilder
		predictor *FirstOccurrencePredictor
	}
)

func NewProfilingWorker(jobConfig *view.JobConfig) (IWorker, error) {
	switch jobConfig.ProfileConfig.ProfileType {
	case profileTypeFirstOccurrence:
		return NewFirstOccurrenceWorker(jobConfig, repository.MongoRepository{})
	default:
		return nil, fmt.Errorf("cant find profile with type %v", jobConfig.ProfileConfig.ProfileType)
	}
}

func NewFirstOccurrenceWorker(jobConfig *view.JobConfig, repo repository.IRepository) (*FirstOccurrenceWorker, error) {
	builderCfg := map[string]interface{}{
		"bootstrap.servers": jobConfig.LogSourceBuilderConfig.Config.BootstrapServer,
		"topics":            jobConfig.LogSourceBuilderConfig.Config.Topic,
		"group.id":          fmt.Sprintf("%s_%v", config.GlobalConfig.KafkaLogsourceGroupIDPrefix, jobConfig.ID),
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
		"bootstrap.servers": jobConfig.LogSourcePredictorConfig.Config.BootstrapServer,
		"topics":            jobConfig.LogSourcePredictorConfig.Config.Topic,
		"group.id":          fmt.Sprintf("%s_%v", config.GlobalConfig.KafkaLogsourceGroupIDPrefix, jobConfig.ID),
	}
	predictorConsumer, err := external.FactoryConsumer(predictorCfg)
	if err != nil {
		return nil, err
	}
	predictor, err := NewFirstOccurrencePredictor(
		jobConfig.ID,
		jobConfig.ProfileConfig,
		predictorConsumer.RcvChannel(),
		builder.GetCommunicationChan())
	if err != nil {
		return nil, err
	}

	builderConsumer.Start()
	predictorConsumer.Start()
	return &FirstOccurrenceWorker{
		id:        jobConfig.ID,
		builder:   builder,
		predictor: predictor,
	}, nil
}

func (w *FirstOccurrenceWorker) Start() {
	w.builder.Start()
	w.predictor.Start()
}

func (w *FirstOccurrenceWorker) Stop() {
	w.builder.Stop()
	w.predictor.Stop()
}
