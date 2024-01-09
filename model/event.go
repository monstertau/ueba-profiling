package model

type Event struct {
	RawEvent   []byte
	Entities   string
	Attributes string
	Result     float64
	Threshold  float64
}

func (e *Event) Marshal() []byte {
	return e.RawEvent
}
