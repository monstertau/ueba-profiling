package external

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
	"sync/atomic"
	"time"
	"ueba-profiling/config"
	"ueba-profiling/util"

	"github.com/IBM/sarama"
)

// SaramaConsumer x
type SaramaConsumer struct {
	consumers   []sarama.ConsumerGroup
	rcvChn      chan []byte
	IsRunning   bool
	numMsg      uint32
	topics      []string
	managerChan chan interface{}
	logger      *logrus.Entry
}

// ConsumerHandler represents a Sarama consumer group consumer
type ConsumerHandler struct {
	Ready  chan bool
	Reader *SaramaConsumer
}

func NewSaramaConsumer(cnf map[string]interface{}) (KafkaConsumer, error) {
	logrus.Infof("Topics string: %s", cnf["topics"].([]interface{}))
	bootstrap := strings.Split(cnf["bootstrap.servers"].(string), ",")
	topics := make([]string, len(cnf["topics"].([]interface{})))
	for i, t := range cnf["topics"].([]interface{}) {
		topics[i] = t.(string)
	}
	groupId := cnf["group.id"].(string)

	numWorkers := config.GlobalConfig.NumConsumer
	consumers := make([]sarama.ConsumerGroup, 0, numWorkers)
	for i := 0; i < numWorkers; i++ {
		cfg := sarama.NewConfig()
		cfg.ChannelBufferSize = 16
		if consumer, err := sarama.NewConsumerGroup(bootstrap, groupId, cfg); err == nil {
			consumers = append(consumers, consumer)
		} else {
			return nil, errors.New("Error creating consumer client, " + err.Error())
		}
	}

	reader := &SaramaConsumer{
		consumers:   consumers,
		rcvChn:      make(chan []byte, 1000),
		IsRunning:   true,
		managerChan: make(chan interface{}),
		numMsg:      0,
		topics:      topics,
		logger:      logrus.WithField("KafkaConsumer", fmt.Sprintf("%s/%s/%s", bootstrap, groupId, topics)),
	}
	return reader, nil
}

func (k *SaramaConsumer) Start() {
	consumerHandlers := make([]*ConsumerHandler, 0, len(k.consumers))

	for _, consumer := range k.consumers {
		consumerHandler := &ConsumerHandler{
			Ready:  make(chan bool),
			Reader: k,
		}
		consumerHandlers = append(consumerHandlers, consumerHandler)
		go func(consumer sarama.ConsumerGroup, handler *ConsumerHandler) {
			for {
				if err := consumer.Consume(context.Background(), k.topics, handler); err != nil {
					k.logger.Errorf("error in consume msg")
				}
				if !k.IsRunning {
					return
				}
				(*handler).Ready = make(chan bool)
			}
		}(consumer, consumerHandler)
	}
	for _, consumerHandler := range consumerHandlers {
		<-consumerHandler.Ready
	}
	k.logger.Infof("Kafka Consumer start!")
	go k.monitor()
}

func (k *SaramaConsumer) monitor() {
	ticker := time.NewTicker(30 * time.Second)
loop:
	for {
		select {
		case <-ticker.C:
			numMsg := atomic.SwapUint32(&k.numMsg, 0)
			k.logger.Infof("consumer consume %d msg", numMsg)
		case ctrlMsg := <-k.managerChan:
			switch ctrlMsg := ctrlMsg.(type) {
			case *util.CloseCtrMsg:
				k.close()
				ctrlMsg.Done()
				k.logger.Infof("consumer closed")
				break loop
			}
		}
	}
}

func (k *SaramaConsumer) RcvChannel() chan []byte {
	return k.rcvChn
}

func (k *SaramaConsumer) Stop() {
	ctrlMsg := util.NewCloseCtrlMsg()
	k.managerChan <- ctrlMsg
	ctrlMsg.WaitForDone()
}

func (k *SaramaConsumer) close() {
	k.IsRunning = false
	for _, r := range k.consumers {
		_ = r.Close()
	}
	close(k.rcvChn)
	close(k.managerChan)

}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *ConsumerHandler) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.Ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *ConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgChan := claim.Messages()
	for {
		select {
		case m, ok := <-msgChan:
			if !ok {
				fmt.Println("kafkaMessages is closed, stopping consumerGroupHandler")
				return nil
			}
			rawMsg := make([]byte, len(m.Value))
			copy(rawMsg, m.Value)
			consumer.Reader.rcvChn <- rawMsg
			atomic.AddUint32(&consumer.Reader.numMsg, 1)
			session.MarkMessage(m, "")
		case <-session.Context().Done():
			return nil
		}
	}
}
