package domain

import "time"

type Telemetry struct {
	Sensor    SensorName
	Value     Value
	Timestamp Timestamp
}

func NewTelemetry(sensor string, value float64, ts time.Time) (Telemetry, error) {
	name, err := NewSensorName(sensor)
	if err != nil {
		return Telemetry{}, err
	}

	return Telemetry{
		Sensor:    name,
		Value:     NewValue(value),
		Timestamp: NewTimestamp(ts),
	}, nil
}
