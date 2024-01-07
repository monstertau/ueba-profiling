package external

import (
	"errors"
	"fmt"
	"ueba-profiling/config"
)

type (
	KafkaConsumer interface {
		Start()
		RcvChannel() chan []byte
		Stop()
	}
	NewConsumerFunc = func(cnf map[string]interface{}) (KafkaConsumer, error)
)

var consumerMap = map[string]NewConsumerFunc{
	config.KafkaTypeSegmentio: NewSegmentioConsumer,
	config.KafkaTypeSarama:    NewSaramaConsumer,
}

func FactoryConsumer(cnf map[string]interface{}) (KafkaConsumer, error) {
	kType := config.GlobalConfig.ConsumerType
	f, ok := consumerMap[kType]
	if !ok {
		return nil, errors.New(fmt.Sprintf("undefined consumer type: %v", kType))
	}
	consumer, err := f(cnf)
	if err != nil {
		return nil, err
	}
	return consumer, nil
}
