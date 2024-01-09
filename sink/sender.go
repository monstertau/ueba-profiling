package sink

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"time"
	"ueba-profiling/util"
	"ueba-profiling/view"
)

const (
	KafkaOutputType = "kafka"
	RedisOutputType = "redis"
)

type (
	Sender interface {
		Send([]interface{}) error
	}
	GoKafka struct {
		writer  IWriter
		logger  *logrus.Entry
		timeout time.Duration
	}
	IWriter interface {
		WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	}
)

func FactorySender(conf *view.OutputConfig) (Sender, error) {
	switch conf.Type {

	case KafkaOutputType:
		kafkaClient, err := NewGoKafka(conf.Config)
		if err != nil {
			return nil, errors.Wrap(err, "new kafka sender")
		}
		return kafkaClient, nil
	default:
		return nil, errors.Errorf("not supported sender type %s", conf.Type)
	}
}

func NewGoKafka(c map[string]interface{}) (*GoKafka, error) {
	bootstrapServers := c["bootstrap.servers"].(string)
	topics := c["topics"].(string)
	writer := &kafka.Writer{
		Addr:     kafka.TCP(bootstrapServers),
		Topic:    topics,
		Balancer: &kafka.RoundRobin{},
		Async:    true,
	}
	conn, err := kafka.DialLeader(context.Background(), "tcp", bootstrapServers, topics, 0)
	if err != nil {
		panic(err)
	}
	// close the connection because we won't be using it
	conn.Close()
	var timeout time.Duration
	timeoutStr, ok := c["producer_timeout"].(string)
	if !ok {
		timeout = 30 * time.Second
	}
	timeout, err = util.ParseDurationExtended(timeoutStr)
	if err != nil {
		logrus.Warnf("error in parse producer_timeout: %v", err)
		timeout = 30 * time.Second
	}
	return &GoKafka{
		writer:  writer,
		logger:  logrus.WithField("kafka_producer", fmt.Sprintf("%v/%v", bootstrapServers, topics)),
		timeout: timeout,
	}, nil
}

func (g *GoKafka) Send(values []interface{}) error {
	if len(values) == 0 {
		return nil
	}
	var messages []kafka.Message
	for _, v := range values {
		b := v.([]byte)
		msg := kafka.Message{Value: b}
		messages = append(messages, msg)
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	err1 := g.writer.WriteMessages(ctx, messages...)
	if err1 != nil {
		var errorMsg string
		switch err := err1.(type) {
		case nil:
		case kafka.WriteErrors:
			errorMsg += fmt.Sprintf("kafka write errors (%d/%d)", err.Count(), len(err)) + ":\n"
			for i := range messages {
				if err[i] != nil {
					// handle the error writing msgs[i]
					errorMsg += err[i].Error() + "\n"
				}
			}
		default:
			return errors.Wrap(err, "WriteMessages")
		}
		return errors.New("WriteMessages: " + errorMsg)
	}

	return nil
}
