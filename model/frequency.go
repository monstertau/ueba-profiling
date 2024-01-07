package model

import "time"

type Frequency struct {
	ID         string    `bson:"id"`
	StartTime  time.Time `bson:"start_time"`
	EndTime    time.Time `bson:"end_time"`
	Entities   string    `bson:"entities"`
	Attributes string    `bson:"attributes"`
}
