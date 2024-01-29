package external

import (
	"errors"
	"fmt"
)

const (
	KafkaTypeSegmentio = "segmentio"
	KafkaTypeSarama    = "sarama"
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
	KafkaTypeSegmentio: NewSegmentioConsumer,
	KafkaTypeSarama:    NewSaramaConsumer,
}

func FactoryConsumer(cnf map[string]interface{}) (KafkaConsumer, error) {
	kType := KafkaTypeSegmentio
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
