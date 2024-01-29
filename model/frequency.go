package model

import (
	"fmt"
	"strings"
	"time"
)

type Frequency struct {
	ID         string     `json:"id"`
	StartTime  CustomTime `json:"window_start"`
	EndTime    CustomTime `json:"window_end"`
	Entities   string     `json:"entities"`
	Attributes string     `json:"attributes"`
	Count      int        `json:"cnt"`
}

type CustomTime struct {
	time.Time
}

const ctLayout = "2006-01-02 15:04:05"

func (ct *CustomTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		ct.Time = time.Time{}
		return
	}
	ct.Time, err = time.Parse(ctLayout, s)
	return
}

func (ct *CustomTime) MarshalJSON() ([]byte, error) {
	if ct.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", ct.Time.Format(ctLayout))), nil
}
