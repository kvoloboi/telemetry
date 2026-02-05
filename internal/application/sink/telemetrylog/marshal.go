package telemetrylog

import (
	"encoding/binary"
	"math"
	"time"

	"github.com/kvoloboi/telemetry/internal/domain"
)

const (
	sensorLen = 1
	valueLen  = 8
)

func marshal(events []domain.Telemetry) ([]byte, error) {
	var size int
	for _, e := range events {
		size += timestampLen + sensorLen + len(e.Sensor.String()) + valueLen
	}

	buf := make([]byte, 0, size)

	for _, e := range events {
		buf = binary.LittleEndian.AppendUint64(
			buf,
			uint64(e.Timestamp.Time().UnixNano()),
		)

		name := e.Sensor.String()
		buf = append(buf, byte(len(name)))
		buf = append(buf, name...)

		buf = binary.LittleEndian.AppendUint64(
			buf,
			math.Float64bits(e.Value.Float64()),
		)
	}

	return buf, nil
}

func unmarshal(buf []byte) ([]domain.Telemetry, error) {
	var events []domain.Telemetry
	i := 0
	var tmp [8]byte

	for i < len(buf) {
		if i+timestampLen > len(buf) {
			return nil, ErrPartialBatch
		}

		copy(tmp[:], buf[i:i+timestampLen])
		ts := time.Unix(0, int64(binary.LittleEndian.Uint64(tmp[:])))
		i += timestampLen

		if i >= len(buf) {
			return nil, ErrPartialBatch
		}

		nameLen := int(buf[i])
		i += sensorLen
		if i+nameLen > len(buf) {
			return nil, ErrPartialBatch
		}
		sensorName := string(buf[i : i+nameLen])
		i += nameLen

		if i+valueLen > len(buf) {
			return nil, ErrPartialBatch
		}
		copy(tmp[:], buf[i:i+valueLen])
		value := math.Float64frombits(binary.LittleEndian.Uint64(tmp[:]))
		i += valueLen

		event, err := domain.NewTelemetry(sensorName, value, ts)
		if err != nil {
			return nil, err
		}

		events = append(events, event)
	}
	return events, nil
}
