package sink

import (
	"github.com/pkg/errors"
	"time"
	"ueba-profiling/model"
	"ueba-profiling/util"
	"ueba-profiling/view"
)

type SinkWorker struct {
	*util.CommonWorker
	inChan chan *model.Event
	sender Sender
}

func NewSinkWorker(outputConf *view.OutputConfig) (*SinkWorker, error) {
	sender, err := FactorySender(outputConf)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create sender")
	}
	sw := &SinkWorker{
		inChan:       make(chan *model.Event, 1000),
		sender:       sender,
		CommonWorker: util.NewCommonWorker("service", "sink_worker"),
	}
	sw.Start()
	return sw, nil
}

func (d *SinkWorker) Start() {
	go d.send()
}

func (d *SinkWorker) send() {
	validCount, failCount := 0, 0
	ticker := time.NewTicker(30 * time.Second)
	tickerSendMess := time.NewTicker(1000 * time.Millisecond)
	var batch []interface{}
loop:
	for {
		select {
		case v, ok := <-d.inChan:
			if !ok {
				d.inChan = nil
				continue
			}
			batch = append(batch, v.Marshal())

			if len(batch) > 1000 {
				if e := d.sender.Send(batch); e != nil {
					failCount += len(batch)
					d.Logger.Warnf("failed to send message batch to destination: %s", e)
				} else {
					validCount += len(batch)
				}
				batch = make([]interface{}, 0)
			}
			// marshal and append to batch
		case <-ticker.C:
			d.Logger.Infof("Total count: %d. Pass count: %d. Fail count: %d", validCount+failCount, validCount, failCount)
			failCount = 0
			validCount = 0
		case <-tickerSendMess.C:
			if len(batch) == 0 {
				continue
			}
			if e := d.sender.Send(batch); e != nil {
				failCount += len(batch)
				d.Logger.Warnf("failed to send message batch to destination: %s", e)
			} else {
				validCount += len(batch)
			}
			batch = make([]interface{}, 0)
			// send batch message to kafka
		case ctrlMsg := <-d.ManageChan:
			switch ctrlMsg := ctrlMsg.(type) {
			case *util.StopCtrMsg:
				break loop
			case *util.CloseCtrMsg:
				d.close()
				ctrlMsg.Done()
				break loop
			default:
				//ignore
			}
		}
	}
}

func marshalOutputEvent(b []byte) []byte {
	return b
}

func (d *SinkWorker) close() {
	close(d.ManageChan)
}

func (d *SinkWorker) Stop() {
	ctrlMsg := util.NewCloseCtrlMsg()
	d.ManageChan <- ctrlMsg
	ctrlMsg.WaitForDone()
}

func (d *SinkWorker) GetInChan() chan *model.Event {
	return d.inChan
}
