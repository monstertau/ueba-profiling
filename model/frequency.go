package model

import "time"

type Frequency struct {
	ID         string    `json:"id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Entities   string    `json:"entities"`
	Attributes string    `json:"attributes"`
	Count      int       `json:"cnt"`
}
