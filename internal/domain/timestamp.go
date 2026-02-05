package domain

import "time"

type Timestamp struct {
	time time.Time
}

func NewTimestamp(t time.Time) Timestamp {
	return Timestamp{time: t}
}

func (t Timestamp) Time() time.Time {
	return t.time
}
