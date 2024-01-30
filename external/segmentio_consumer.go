package external

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
	"ueba-profiling/util"
)

type NewKafka struct {
	reader      *kafka.Reader
	rcvChn      chan []byte
	numMsg      int
	managerChan chan interface{}
	running     bool
	logger      *logrus.Entry
}

func NewSegmentioConsumer(cnf map[string]interface{}) (KafkaConsumer, error) {
	bootstrap := strings.Split(cnf["bootstrap.servers"].(string), ",")
	topics := strings.Split(cnf["topics"].(string), ",")
	groupId := cnf["group.id"].(string)
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        bootstrap,
		GroupTopics:    topics,
		MaxWait:        time.Second,
		CommitInterval: time.Second,
		GroupID:        groupId,
		StartOffset:    kafka.FirstOffset,
	})
	k := &NewKafka{
		reader:      r,
		rcvChn:      make(chan []byte, 1000),
		managerChan: make(chan interface{}),
		numMsg:      0,
		running:     true,
		logger:      logrus.WithField("KafkaConsumer", fmt.Sprintf("%s/%s/%s", bootstrap, groupId, topics)),
	}
	return k, nil
}
func (k *NewKafka) close() {
	close(k.rcvChn)
	close(k.managerChan)
	k.reader.Close()
}

func (k *NewKafka) Start() {
	ticker := time.NewTicker(5 * time.Second)
	commitTicker := time.NewTicker(time.Second * 10)
	commitMsg := kafka.Message{Offset: -1}
	commit := func(msg kafka.Message) {
		if msg.Offset == -1 {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := k.reader.CommitMessages(ctx, msg); err != nil {
			k.logger.Errorf("failed to commit offset: %s", err)
		}
		k.logger.Infof("committed message at offset %d, partition %d in topic %s", msg.Offset, msg.Partition, msg.Topic)
	}
	fetch := func() (kafka.Message, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		m, err := k.reader.FetchMessage(ctx)
		if err != nil {
			return kafka.Message{}, err
		}
		return m, nil
	}
	go func() {
	loop:
		for {
			select {
			case <-ticker.C:
				k.logger.Infof("consumer consume %d msg", k.numMsg)
				k.numMsg = 0
			case ctrlMsg := <-k.managerChan:
				switch ctrlMsg := ctrlMsg.(type) {
				case *util.CloseCtrMsg:
					k.close()
					ctrlMsg.Done()
					k.logger.Infof("consumer closed")
					break loop
				}
			case <-commitTicker.C:
				commit(commitMsg)
				commitMsg = kafka.Message{Offset: -1}
			default:
				m, err := fetch()
				if err != nil {
					continue
				}
				commitMsg = m
				k.rcvChn <- m.Value
				k.numMsg++
			}
		}
	}()
}

func (k *NewKafka) RcvChannel() chan []byte {
	return k.rcvChn
}

func (k *NewKafka) TestConnection() error {
	_, err := kafka.DialLeader(context.Background(), "tcp", k.reader.Config().Brokers[0], k.reader.Config().Topic, 0)
	if err != nil {
		logrus.Warnf("failed to dial leader:%v", err)
	}
	return err

}

func (k *NewKafka) Stop() {
	closeMsg := util.NewCloseCtrlMsg()
	k.managerChan <- closeMsg
	closeMsg.WaitForDone()
}

func (k *NewKafka) Pause() {

}

func (k *NewKafka) Resume() {

}
