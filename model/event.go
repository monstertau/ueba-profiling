package model

import "encoding/json"

type Event struct {
	RawEvent   []byte
	Entities   string
	Attributes string
	Result     float64
	Threshold  float64
}

func (e *Event) Marshal() []byte {
	event := make(map[string]interface{})
	_ = json.Unmarshal(e.RawEvent, &event)
	event["profile_predictor_entities"] = e.Entities
	event["profile_predictor_attributes"] = e.Attributes
	event["profile_predictor_result"] = e.Result
	event["profile_predictor_threshold"] = e.Threshold
	b, _ := json.Marshal(event)
	return b
}
