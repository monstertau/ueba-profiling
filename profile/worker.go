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
	profileTypeRarity          = "rarity"
)

type (
	IWorker interface {
		Start()
		Stop()
	}
	ProfilingWorker struct {
		id                                 string
		builder                            *ModelBuilder
		builderConsumer, predictorConsumer external.KafkaConsumer
		sinkWorker                         *sink.SinkWorker
		predictor                          IPredictor
	}
	FirstOccurrenceWorker struct {
		predictor *FirstOccurrencePredictor
	}
)

func NewProfilingWorker(jobConfig *view.JobConfig) (IWorker, error) {
	builderCfg := map[string]interface{}{
		"bootstrap.servers": jobConfig.LogSourceBuilderConfig.BootstrapServer,
		"topics":            jobConfig.LogSourceBuilderConfig.Topic,
		"group.id":          fmt.Sprintf("%s_%v", config.AppConfig.KafkaLogSource.GroupID, jobConfig.ID),
	}
	builderConsumer, err := external.FactoryConsumer(builderCfg)
	if err != nil {
		return nil, err
	}
	repo := &repository.MongoRepository{}
	builder, err := NewModelBuilder(jobConfig.ID, builderConsumer.RcvChannel(), jobConfig.ProfileConfig, repo)
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
	predictor, err := FactoryPredictor(
		jobConfig,
		repo,
		predictorConsumer.RcvChannel(),
		sinkWorker.GetInChan(),
		builder.GetCommunicationChan())
	if err != nil {
		return nil, err
	}
	return &ProfilingWorker{
		id:                jobConfig.ID,
		builder:           builder,
		builderConsumer:   builderConsumer,
		predictor:         predictor,
		predictorConsumer: predictorConsumer,
	}, nil
}

func (w *ProfilingWorker) Start() {
	w.builder.Start()
	w.builderConsumer.Start()
	go func() {
		time.Sleep(30 * time.Second)
		w.predictorConsumer.Start()
		w.predictor.Start()
	}()

}

func (w *ProfilingWorker) Stop() {
	w.builderConsumer.Stop()
	w.predictorConsumer.Stop()
	w.builder.Stop()
	w.predictor.Stop()
}
