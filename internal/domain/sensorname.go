package domain

import "errors"

type SensorName struct {
	name string
}

const MaxSensorNameLen = 255

var (
	ErrSensorNameTooLong = errors.New("sensor name too long")
	ErrSensorNameEmpty   = errors.New("sensor name cannot be empty")
)

func NewSensorName(name string) (SensorName, error) {
	if len(name) == 0 {
		return SensorName{}, ErrSensorNameEmpty
	}
	if len(name) > MaxSensorNameLen {
		return SensorName{}, ErrSensorNameTooLong
	}
	return SensorName{name: name}, nil
}

func (s SensorName) String() string {
	return s.name
}
